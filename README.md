# Keno4min Lottery Game Engine

Sistema de lotería tipo Keno con juegos cada 4 minutos, construido con AWS SAM, Lambda (Go) y MongoDB.

## 📋 Arquitectura

- **Lambda Function**: Crear juegos de lotería
- **Step Functions**: Orquestación del workflow de juegos
- **EventBridge**: Scheduler cada 3 minutos
- **Secrets Manager**: Credenciales de MongoDB
- **MongoDB Atlas**: Base de datos

## 🚀 Setup Inicial

### 1. Prerrequisitos

```bash
# Instalar AWS SAM CLI
brew install aws-sam-cli

# Configurar AWS CLI
aws configure

# Instalar Go 1.21+
brew install go
```

### 2. Configuración de Secretos (Primera vez)

Los secretos se crean **una vez** en AWS y nunca se suben a Git:

```bash
# Dar permisos de ejecución al script
chmod +x scripts/setup-secrets.sh

# Crear secreto para DEV
./scripts/setup-secrets.sh dev "mongodb+srv://user:pass@cluster.mongodb.net"

# Crear secreto para STAGING (cuando sea necesario)
./scripts/setup-secrets.sh staging "mongodb+srv://user:pass@staging-cluster.mongodb.net"

# Crear secreto para PROD (cuando sea necesario)
./scripts/setup-secrets.sh prod "mongodb+srv://user:pass@prod-cluster.mongodb.net"
```

**Importante:** Las credenciales quedan en AWS Secrets Manager, no en el código.

### 3. Configuración Local

```bash
# Copiar configuración de ejemplo
cp samconfig.toml.example samconfig.toml

# samconfig.toml ya NO se sube a Git
```

## 🛠️ Desarrollo

### Build

```bash
sam build
```

### Deploy por Entorno

```bash
# Development (default)
sam deploy

# Staging
sam deploy --config-env staging

# Production
sam deploy --config-env prod
```

### Testing Local

```bash
# Invocar la Lambda localmente
sam local invoke CreateGameFunction

# Con evento de prueba
sam local invoke CreateGameFunction -e events/create-game.json
```

## 📁 Estructura del Proyecto

```
.
├── functions/
│   └── create-game/          # Lambda function en Go
│       ├── cmd/lambda/        # Entry point
│       ├── internal/          # Lógica de negocio
│       └── Makefile           # Build de Go
├── statemachine/
│   └── game-workflow.asl.json # Step Functions definition
├── scripts/
│   └── setup-secrets.sh       # Script para crear secretos
├── template.yml               # SAM template (IaC)
├── samconfig.toml             # Config local (NO commitear)
└── .gitignore                 # Protección de secretos
```

## 🔐 Seguridad

### Archivos que NO se suben a Git:
- `samconfig.toml` - Configuración local
- `.secrets.env` - Variables de entorno con credenciales
- `.aws-sam/` - Build artifacts

### Archivos que SÍ se suben a Git:
- `samconfig.toml.example` - Plantilla de configuración
- `.secrets.env.example` - Plantilla de secretos
- `template.yml` - Infraestructura (sin credenciales)

## 🌍 Ambientes

Cada ambiente tiene su propio:
- Stack de CloudFormation
- Secret en Secrets Manager
- Configuración en samconfig.toml

| Ambiente | Stack Name | Secret Path |
|----------|-----------|-------------|
| Dev | `keno4min-lottery-game-engine` | `/keno4min/dev/mongodb` |
| Staging | `keno4min-lottery-game-engine-staging` | `/keno4min/staging/mongodb` |
| Prod | `keno4min-lottery-game-engine-prod` | `/keno4min/prod/mongodb` |

## 🔄 CI/CD (Recomendado para Producción)

Usando GitHub Actions:

```yaml
# .github/workflows/deploy.yml
- name: Deploy to AWS
  env:
    MONGODB_URI: ${{ secrets.MONGODB_URI_PROD }}
  run: |
    ./scripts/setup-secrets.sh prod "$MONGODB_URI"
    sam deploy --config-env prod --no-confirm-changeset
```

## 📝 Comandos Útiles

```bash
# Ver logs en tiempo real
sam logs -n CreateGameFunction --stack-name keno4min-lottery-game-engine --tail

# Ver secretos (solo nombres, no valores)
aws secretsmanager list-secrets --query 'SecretList[?contains(Name, `keno4min`)]'

# Obtener valor de un secreto
aws secretsmanager get-secret-value --secret-id /keno4min/dev/mongodb

# Eliminar stack
sam delete --stack-name keno4min-lottery-game-engine
```

## ❓ FAQ

**Q: ¿Por qué samconfig.toml está en .gitignore?**  
A: Ya no tiene credenciales, pero puede tener configuraciones específicas de tu máquina.

**Q: ¿Cómo compartir credenciales con el equipo?**  
A: NO compartas credenciales. Cada desarrollador ejecuta `setup-secrets.sh` con sus propias credenciales.

**Q: ¿Cómo rotar credenciales?**  
A: Ejecuta nuevamente `setup-secrets.sh` con las nuevas credenciales. El script actualiza el secreto existente.

## 📄 Licencia

[Tu licencia aquí]
