# Keno4min Draw Engine

Proyecto AWS SAM que crea sorteos en una tabla DynamoDB existente.

```text
EventBridge Scheduler (cada 3 minutos)
  └── GameWorkflow (Step Functions STANDARD)
        ├── CreateDrawFunction (Go, provided.al2023, ARM64)
        └── WaitUntilClosesAt (TimestampPath: $.closesAt)
```

El workflow entrega el resultado de la Lambda al estado Wait y termina al llegar
a `closesAt`. Los pasos `draw → update_tickets` todavía están pendientes.
La espera ocurre en Step Functions, sin mantener una Lambda ejecutándose.
Si la fecha ya pasó, continúa inmediatamente. El estado Wait tiene precisión de
segundos y trunca las fracciones del timestamp.

## Código

```text
functions/create-draw/
├── cmd/lambda/main.go              # Configuración, dependencias y handler
├── internal/draw/
│   ├── draw.go                     # Draw y ventana de 157 segundos
│   └── create.go                   # Caso de uso y contrato Repository
├── internal/dynamodb/
│   ├── counter.go                  # NextTIR con incremento atómico
│   ├── repository.go               # Consulta por requestId y guardado
│   └── item.go                     # Formato persistido y conversión
├── Makefile                        # Binario bootstrap para Linux ARM64
├── go.mod
└── go.sum
```

Las pruebas viven junto al código. `draw` depende únicamente de la biblioteca
estándar, incluido el paquete `uuid`; `dynamodb` implementa su contrato. `main.go` construye el cliente
AWS una vez y adapta la entrada de Lambda. No se necesita un contenedor de inyección.

`Create.Execute` sigue cuatro pasos:

1. Consultar si `requestId` ya tiene un draw y devolverlo si existe.
2. Obtener el siguiente TIR con `NextTIR`.
3. Construir un draw abierto, con UUID y fechas UTC.
4. Guardar el draw y sus referencias en una transacción.

## Contador e idempotencia

`NextTIR` sigue el patrón de `GenerateCoupon`: usa `UpdateItem` con
`ADD #value :one` y `ReturnValues: UPDATED_NEW`. No necesita leer el contador
antes de incrementarlo ni mantener una entidad `DrawCounter`.

El contador está en `COUNTER#DRAW_TIR / META`, atributo Number `value`.
Si no existe comienza en 1, conservando el comportamiento anterior.
Un contador existente debe tener un entero no negativo y no estar agotado;
un valor inválido devuelve error. Usa el rango de `uint64`, hasta
18.446.744.073.709.551.615 (`math.MaxUint64`), y rechaza nuevos incrementos al agotarse.

**Si la tabla ya contiene sorteos, inicializar el contador con el último TIR
antes de habilitar el schedule.** La función no escanea ni migra datos.

El incremento ocurre antes del guardado y no se revierte si falla la creación.
Por tanto, los TIR son únicos y crecientes, pero pueden tener saltos por fallos,
concurrencia o reintentos del SDK. Es el comportamiento esperado de un
[contador atómico de DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/WorkingWithItems.html#WorkingWithItems.AtomicCounters).
Esta es la diferencia frente a la versión TypeScript, que incluía el contador
en la transacción y evitaba consumir TIR cuando se cancelaba el guardado.

Un `requestId` confirmado siempre devuelve el mismo draw sin incrementar el
contador ni volver a publicar `current`. Las entregas simultáneas pueden consumir
varios TIR, pero la condición de no existencia de `DRAW_REQUEST` permite guardar
un solo draw. Step Functions reintenta `DrawConflictError` hasta cinco veces
con espera exponencial, consultando de nuevo `requestId` en cada invocación.
Los registros de idempotencia no tienen TTL.

## Persistencia

El contador usa una operación independiente. El guardado reúne estos cuatro cambios
en un único `TransactWriteItems`:

| PK | SK | Cambio |
|---|---|---|
| `DRAW#<id>` | `META` | Crea el draw sin sobrescribir |
| `DRAW_TIR#<tir>` | `META` | Reserva la referencia `id` sin sobrescribir |
| `DRAW_RUNTIME` | `META` | Reemplaza `current` y conserva `latestResult` |
| `DRAW_REQUEST#<requestId>` | `META` | Registra `id` y `tir` sin sobrescribir |

`current` solo se reemplaza si no existe o tiene un TIR menor. Una asignación
atrasada no puede hacer retroceder el runtime; se cancela toda la transacción y
el workflow puede reintentar con otro TIR. Los otros errores del SDK se propagan.

Se conservan los nombres de atributos existentes (`id`, `tir`, `status`,
`startsAt`, `closesAt`, `balls`, `videoId`) y las claves DynamoDB. Las fechas se
persisten como strings ISO-8601 UTC, compatibles con los registros TypeScript.
Al crear no se guardan `balls`, `videoId`, `GSI1PK` ni `GSI1SK`.
Las lecturas por `requestId` y del draw son fuertemente consistentes.

## Configuración y contrato

Requisitos: Go 1.27.3, AWS SAM CLI y Docker para las pruebas de integración.
El módulo fija Go 1.27.3 como versión mínima, la última estable al actualizarlo;
Go descarga el toolchain necesario cuando la instalación local es anterior.
Los UUID v4 se generan con `uuid.NewV4()` de la biblioteca estándar, sin dependencias externas.
`sam build` ejecuta el Makefile y genera `bootstrap` para Linux ARM64, con
`CGO_ENABLED=0` y `lambda.norpc`.

`DynamoDBTableName` configura `DYNAMODB_TABLE` en Lambda. La tabla debe existir
en la misma cuenta y región, con claves String `PK` y `SK`. No se necesitan
índices para crear un draw. SAM concede permisos GetItem, PutItem y UpdateItem
sobre esa tabla, incluidas las escrituras transaccionales.

Entrada:

```json
{"requestId":"local-create-draw-001"}
```

`requestId` es obligatorio. El schedule combina su ARN y la hora programada para
mantenerlo entre intentos. Para una creación manual nueva hay que usar otro valor.

Salida:

```json
{
  "id": "ef5ebed8-7871-4fa4-bfe4-f5dc0bb139ba",
  "tir": 1,
  "status": "open",
  "startsAt": "2026-09-12T12:00:00Z",
  "closesAt": "2026-09-12T12:02:37Z"
}
```

`closesAt = startsAt + 157 segundos`. Las fechas se convierten a UTC sin truncar
su precisión. La idempotencia cubre la creación; las futuras etapas del workflow
deberán implementar su propia protección frente a duplicados.

## Validación

Desde `functions/create-draw/`:

```sh
go test -race ./...
go vet ./...
```

Desde la raíz:

```sh
sam validate --lint --template-file template.yml
sam build --template-file template.yml
```

## Integración local

```sh
docker run --rm -d --name keno4min-create-draw-test \
  -p 127.0.0.1:18073:8000 amazon/dynamodb-local \
  -jar DynamoDBLocal.jar -inMemory -sharedDb
```

Desde `functions/create-draw/`:

```sh
DYNAMODB_INTEGRATION=1 go test -race ./internal/dynamodb -run TestDynamoDBIntegration -count=1 -v
docker stop keno4min-create-draw-test
```

Sin `DYNAMODB_INTEGRATION=1`, se omite la integración. La prueba solo conecta a
`127.0.0.1:18073` con credenciales ficticias, crea una tabla temporal y la elimina.
Verifica idempotencia, concurrencia, TIR únicos, conservación de `latestResult`,
rollback del guardado, contadores inválidos y protección contra retrocesos del runtime.

## Invocación y despliegue

```sh
sam local invoke CreateDrawFunction --event events/create-draw.json
```

La invocación SAM usa las credenciales AWS disponibles y **escribe en la tabla
configurada**. Para pruebas sin AWS, usar la integración local anterior.

`samconfig.toml` define dos stacks independientes:

| Ambiente | Stack | Tabla | Schedule | Región |
|---|---|---|---|---|
| dev | keno4min-draw-engine-dev | cronos-dev-keno4min | ENABLED | us-east-1 |
| prod | keno4min-draw-engine-prod | cronos-prod-keno4min | ENABLED | us-east-1 |

La tabla de producción debe existir antes del despliegue.

```sh
sam build --template-file template.yml
sam deploy --config-env dev
# Producción:
sam deploy --config-env prod
```

El flujo de configuración es `samconfig.toml → DynamoDBTableName → DYNAMODB_TABLE`.
No es necesario cambiar el código entre ambientes. `--config-env` selecciona
configuración SAM, no cuenta ni credenciales. Añadir `--profile <perfil>` cuando
corresponda; no se almacenan credenciales en el repositorio.

Para pausar nuevas ejecuciones, cambiar `ScheduleState=DISABLED` en los
`parameter_overrides` del ambiente y desplegarlo. Las ejecuciones iniciadas continúan.
Los valores de `deploy` no se aplican a `sam local invoke`, que usa dev por defecto.
Para otra tabla local, pasar `--parameter-overrides DynamoDBTableName=<tabla>`.
