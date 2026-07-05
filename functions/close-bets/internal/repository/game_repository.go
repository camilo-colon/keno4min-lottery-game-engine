package repository

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GameRepository struct {
	collection *mongo.Collection
}

func NewGameRepository(db *mongo.Database) domain.GameRepository {
	return &GameRepository{
		collection: db.Collection("games"),
	}
}

// UpdateStatus actualiza el estado de un juego
func (r *GameRepository) UpdateStatus(ctx context.Context, gameID string, from, to domain.GameStatus) error {
	filter := bson.M{
		"_id":    gameID,
		"status": from,
	}
	update := bson.M{
		"$set": bson.M{
			"status": to,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating game status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("game %s not found with status %s", gameID, from)
	}

	return nil
}
