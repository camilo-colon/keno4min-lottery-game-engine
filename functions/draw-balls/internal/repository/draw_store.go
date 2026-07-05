package repository

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const keno4MinGameID = "keno4min"
const drawsDatabaseName = "lottery"

type DrawStore struct {
	collection *mongo.Collection
}

func NewDrawStore(client *mongo.Client) *DrawStore {
	return &DrawStore{
		collection: client.Database(drawsDatabaseName).Collection("draws"),
	}
}

func (s *DrawStore) GetRandomKeno4MinDraw(ctx context.Context) (*domain.Draws, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"game": keno4MinGameID}}},
		bson.D{{Key: "$sample", Value: bson.M{"size": 1}}},
	}

	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get random draw: %w", err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, fmt.Errorf("no draws available for game %s", keno4MinGameID)
	}

	var draw domain.Draws
	if err := cursor.Decode(&draw); err != nil {
		return nil, fmt.Errorf("failed to decode draw: %w", err)
	}

	return &draw, nil
}
