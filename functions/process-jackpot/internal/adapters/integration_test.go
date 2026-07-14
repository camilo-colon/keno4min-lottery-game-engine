//go:build integration

package adapters_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/adapters"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/ports"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	testClient *mongo.Client
	testDB     *mongo.Database
)

// TestMain levanta un MongoDB real en modo replica set (necesario para
// transacciones) una sola vez para toda la suite de integración.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:7", mongodb.WithReplicaSet("rs0"))
	if err != nil {
		panic("start mongodb container: " + err.Error())
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		panic("connection string: " + err.Error())
	}

	// El RS de un solo nodo se inicia anunciando la IP interna del contenedor,
	// inalcanzable desde el host. Conectamos con directConnection=true (sin
	// replicaSet) para hablar directo al puerto mapeado sin descubrir topología.
	// Las transacciones siguen soportadas: el nodo es primary de un RS.
	u, err := url.Parse(uri)
	if err != nil {
		panic("parse uri: " + err.Error())
	}
	q := u.Query()
	q.Del("replicaSet")
	q.Set("directConnection", "true")
	u.RawQuery = q.Encode()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
	if err != nil {
		panic("connect mongo: " + err.Error())
	}
	testClient = client
	testDB = client.Database("lottery_test")

	code := m.Run()

	_ = client.Disconnect(ctx)
	_ = container.Terminate(ctx)
	os.Exit(code)
}

// createRunsUniqueIndex crea el índice único que en producción se crea a mano
// (migración) y que actúa como guardián de idempotencia.
func createRunsUniqueIndex(ctx context.Context, t *testing.T) {
	t.Helper()
	_, err := testDB.Collection("jackpot_runs").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "game_id", Value: 1}, {Key: "club_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_game_club"),
	})
	if err != nil {
		t.Fatalf("crear índice único: %v", err)
	}
}

// TestTicketStoreExcludesCanceled verifica que el filtro del repositorio deja
// fuera a los tickets cancelados (la garantía que perdió el test unitario).
func TestTicketStoreExcludesCanceled(t *testing.T) {
	ctx := context.Background()
	col := testDB.Collection("tickets")
	if err := col.Drop(ctx); err != nil {
		t.Fatal(err)
	}

	_, err := col.InsertMany(ctx, []any{
		bson.M{"_id": "t1", "game_id": "g1", "club_id": "A", "state": "PAYED"},
		bson.M{"_id": "t2", "game_id": "g1", "club_id": "A", "state": "CANCELED"},
		bson.M{"_id": "t3", "game_id": "g1", "club_id": "A", "state": "LOSS"},
		bson.M{"_id": "t4", "game_id": "g1", "club_id": "B", "state": "CANCELED"}, // club B: solo cancelado
	})
	if err != nil {
		t.Fatal(err)
	}

	store := adapters.NewTicketStore(testDB)

	// FindClubIDsByGame: B no aparece (solo tiene cancelados).
	clubIDs, err := store.FindClubIDsByGame(ctx, "g1")
	if err != nil {
		t.Fatal(err)
	}
	if len(clubIDs) != 1 || clubIDs[0] != "A" {
		t.Errorf("clubIDs = %v, want [A]", clubIDs)
	}

	// FindByClubAndGame: excluye el cancelado t2.
	tickets, err := store.FindByClubAndGame(ctx, "A", "g1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tickets) != 2 {
		t.Fatalf("len(tickets) = %d, want 2 (excluye cancelado)", len(tickets))
	}
	for _, tk := range tickets {
		if tk.State == domain.CANCELED {
			t.Errorf("no debía venir el ticket cancelado %s", tk.ID)
		}
	}
}

// TestRunStoreIdempotency verifica el guardián de idempotencia vía índice único.
func TestRunStoreIdempotency(t *testing.T) {
	ctx := context.Background()
	if err := testDB.Collection("jackpot_runs").Drop(ctx); err != nil {
		t.Fatal(err)
	}

	createRunsUniqueIndex(ctx, t)
	store := adapters.NewRunStore(testDB)

	if err := store.Mark(ctx, "g1", "A"); err != nil {
		t.Fatalf("primera marca debe pasar: %v", err)
	}
	if err := store.Mark(ctx, "g1", "A"); !errors.Is(err, ports.ErrAlreadyProcessed) {
		t.Fatalf("segunda marca del mismo par debe ser ErrAlreadyProcessed, got %v", err)
	}
	// Pares distintos sí pasan.
	if err := store.Mark(ctx, "g1", "B"); err != nil {
		t.Fatalf("otro club debe pasar: %v", err)
	}
	if err := store.Mark(ctx, "g2", "A"); err != nil {
		t.Fatalf("otro game debe pasar: %v", err)
	}
}

// TestTransactionRollback verifica que un error dentro de la transacción hace
// rollback REAL: la marca escrita NO queda persistida.
func TestTransactionRollback(t *testing.T) {
	ctx := context.Background()
	runs := testDB.Collection("jackpot_runs")
	if err := runs.Drop(ctx); err != nil {
		t.Fatal(err)
	}
	createRunsUniqueIndex(ctx, t)
	store := adapters.NewRunStore(testDB)

	tm := adapters.NewMongoTransactionManager(testClient)

	boom := errors.New("boom")
	err := tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := store.Mark(txCtx, "g1", "A"); err != nil {
			return err
		}
		return boom // fuerza el rollback tras escribir la marca
	})
	if !errors.Is(err, boom) {
		t.Fatalf("se esperaba el error boom, got %v", err)
	}

	count, err := runs.CountDocuments(ctx, bson.M{"game_id": "g1", "club_id": "A"})
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("rollback fallido: quedaron %d marcas, want 0", count)
	}
}

// TestClubStoreIncrementAndReset verifica el $inc atómico y el reemplazo de jp1.
func TestClubStoreIncrementAndReset(t *testing.T) {
	ctx := context.Background()
	clubs := testDB.Collection("clubs")
	if err := clubs.Drop(ctx); err != nil {
		t.Fatal(err)
	}

	_, err := clubs.InsertOne(ctx, bson.M{
		"_id":    "A",
		"jp1":    bson.M{"percent": 1, "target": 100, "value": 50, "min": 15, "max": 25},
		"config": bson.M{"jackpot": bson.M{"min": 15, "max": 25, "percent": 3, "seed": 5}},
	})
	if err != nil {
		t.Fatal(err)
	}

	store := adapters.NewClubStore(testDB)

	// $inc atómico devuelve el valor ya incrementado.
	jp, err := store.IncrementJackpot(ctx, "A", 30)
	if err != nil {
		t.Fatal(err)
	}
	if jp.Value != 80 {
		t.Errorf("value tras incremento = %d, want 80", jp.Value)
	}

	// Reset: jp1 queda con la config y sin _id residual.
	fresh := domain.NewJackpotFromConfig(domain.JackpotConfig{Min: 15, Max: 25, Percent: 3, Seed: 5}, 20)
	if err := store.ResetJackpot(ctx, "A", fresh); err != nil {
		t.Fatal(err)
	}

	reloaded, err := store.FindByID(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.JP1.Value != 5 || reloaded.JP1.Target != 20 || reloaded.JP1.Percent != 3 {
		t.Errorf("jp1 tras reset = %+v, want value=5 target=20 percent=3", reloaded.JP1)
	}
}
