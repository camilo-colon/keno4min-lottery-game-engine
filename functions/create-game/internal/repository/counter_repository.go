package repository

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/domain"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CounterRepository struct {
	collection *mongo.Collection
}

func NewCounterRepository(db *mongo.Database) *CounterRepository {
	return &CounterRepository{
		collection: db.Collection("counters"),
	}
}

// IncrementAndGet incrementa el contador y devuelve el nuevo valor
func (r *CounterRepository) IncrementAndGet(ctx context.Context) (uint64, error) {
	filter := map[string]any{"_id": "games_round"}
	update := map[string]any{
		"$inc": map[string]any{"value": 1},
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var counter domain.Counter
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, err
	}

	return counter.Count, nil
}
