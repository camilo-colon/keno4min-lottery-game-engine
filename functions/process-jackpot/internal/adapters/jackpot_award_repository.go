package adapters

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

// JackpotAwardStore persiste el histórico de jackpots ganados en la colección
// jackpot_awards.
//
// Índices recomendados (creados manualmente, sin TTL: es historia permanente):
//
//	db.jackpot_awards.createIndex({ club_id: 1, created_at: -1 })
//	db.jackpot_awards.createIndex({ game_id: 1 })
type JackpotAwardStore struct {
	collection *mongo.Collection
}

func NewJackpotAwardStore(db *mongo.Database) *JackpotAwardStore {
	return &JackpotAwardStore{
		collection: db.Collection("jackpot_awards"),
	}
}

// Record guarda un premio otorgado.
func (r *JackpotAwardStore) Record(ctx context.Context, award domain.JackpotAward) error {
	if _, err := r.collection.InsertOne(ctx, award); err != nil {
		return fmt.Errorf("error recording jackpot award: %w", err)
	}
	return nil
}
