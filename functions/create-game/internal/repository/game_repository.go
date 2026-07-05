package repository

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

type GameRepository struct {
	collection *mongo.Collection
}

func NewGameRepository(db *mongo.Database) *GameRepository {
	return &GameRepository{
		collection: db.Collection("games"),
	}
}

func (r *GameRepository) Create(ctx context.Context, game *domain.Game) error {
	_, err := r.collection.InsertOne(ctx, game)
	return err
}
