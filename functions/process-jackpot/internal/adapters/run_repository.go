package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RunStore registra el procesamiento del jackpot por (game, club).
//
// REQUIERE un índice único creado manualmente (migración) sobre la colección
// jackpot_runs; ese índice es el guardián de idempotencia:
//
//	db.jackpot_runs.createIndex({ game_id: 1, club_id: 1 }, { unique: true, name: "uniq_game_club" })
//
// Recomendado además un TTL para autolimpiar marcas viejas:
//
//	db.jackpot_runs.createIndex({ created_at: 1 }, { expireAfterSeconds: 604800, name: "ttl_created_at" })
type RunStore struct {
	collection *mongo.Collection
}

func NewRunStore(db *mongo.Database) *RunStore {
	return &RunStore{
		collection: db.Collection("jackpot_runs"),
	}
}

// Mark registra que (gameID, clubID) fue procesado. Si el índice único rechaza el
// insert (ya existía), devuelve ports.ErrAlreadyProcessed.
func (r *RunStore) Mark(ctx context.Context, gameID, clubID string) error {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"game_id":    gameID,
		"club_id":    clubID,
		"created_at": time.Now().UTC(),
	})
	if mongo.IsDuplicateKeyError(err) {
		return ports.ErrAlreadyProcessed
	}
	if err != nil {
		return fmt.Errorf("error marking jackpot run for game %s club %s: %w", gameID, clubID, err)
	}
	return nil
}
