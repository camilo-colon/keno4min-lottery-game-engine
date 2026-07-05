package repository

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TicketStore struct {
	collection *mongo.Collection
}

func NewTicketStore(db *mongo.Database) *TicketStore {
	return &TicketStore{
		collection: db.Collection("tickets"),
	}
}

func (s *TicketStore) GetStats(ctx context.Context) (*domain.Stats, error) {
	pipeline := bson.A{
		bson.M{
			"$group": bson.M{
				"_id":          nil,
				"total_income": bson.M{"$sum": "$total"},
				"total_paid":   bson.M{"$sum": "$win"},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":          0,
				"total_income": 1,
				"total_paid":   1,
				"rtp": bson.M{"$cond": bson.A{
					bson.M{"$gt": bson.A{"$total_income", 0}},
					bson.M{"$multiply": bson.A{bson.M{"$divide": bson.A{"$total_paid", "$total_income"}}, 100}},
					0,
				}},
			},
		},
	}

	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result domain.Stats
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	}
	return &result, nil
}

func (s *TicketStore) GetTicketsByGame(ctx context.Context, gameID string) ([]domain.Ticket, error) {
	filter := bson.M{"game_id": gameID}
	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var tickets []domain.Ticket
	if err := cursor.All(ctx, &tickets); err != nil {
		return nil, err
	}
	return tickets, nil
}
