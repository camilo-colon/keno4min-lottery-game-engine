package repository

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GameStore struct {
	collection *mongo.Collection
}

func NewGameStore(db *mongo.Database) *GameStore {
	return &GameStore{
		collection: db.Collection("games"),
	}
}

func (r *GameStore) GetHistoryGame(ctx context.Context, limit int64) ([]domain.Game, error) {
	filter := bson.M{
		"status": "DRAWN",
	}
	options := options.Find().SetSort(bson.D{{Key: "round", Value: -1}}).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, filter, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get game history: %w", err)
	}
	defer cursor.Close(ctx)

	var games []domain.Game
	if err := cursor.All(ctx, &games); err != nil {
		return nil, fmt.Errorf("failed to decode game history: %w", err)
	}
	return games, nil
}

// FindByID busca un juego por su ID
func (r *GameStore) FindByID(ctx context.Context, gameID string) (*domain.Game, error) {
	var game domain.Game
	err := r.collection.FindOne(ctx, bson.M{"_id": gameID}).Decode(&game)
	if err != nil {
		return nil, fmt.Errorf("game %s not found: %w", gameID, err)
	}
	return &game, nil
}

func (r *GameStore) UpdateGame(ctx context.Context, game *domain.Game) error {
	filter := bson.M{"_id": game.ID}
	update := bson.M{
		"$set": bson.M{
			"status": game.Status,
			"idv":    game.Idv,
			"balls":  game.Balls,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error updating game: %w", err)
	}

	return nil
}
