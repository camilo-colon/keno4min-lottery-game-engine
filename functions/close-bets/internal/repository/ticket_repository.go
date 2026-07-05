package repository

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TicketRepository struct {
	collection *mongo.Collection
}

func NewTicketRepository(db *mongo.Database) domain.TicketRepository {
	return &TicketRepository{
		collection: db.Collection("tickets"),
	}
}

func (r *TicketRepository) GetTicketsByGameID(ctx context.Context, gameID string, cursor *string, limit int64) ([]domain.Ticket, *string, error) {
	if limit <= 0 {
		limit = 100
	}

	filter := bson.M{
		"game_id": gameID,
		"state":   domain.TicketStatePending,
	}

	if cursor != nil && *cursor != "" {
		filter["_id"] = bson.M{"$gt": *cursor}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(limit)

	mongoCursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("error finding tickets for game %s: %w", gameID, err)
	}
	defer mongoCursor.Close(ctx)

	var tickets []domain.Ticket
	if err := mongoCursor.All(ctx, &tickets); err != nil {
		return nil, nil, fmt.Errorf("error decoding tickets for game %s: %w", gameID, err)
	}

	if len(tickets) == 0 {
		return tickets, nil, nil
	}

	nextCursor := tickets[len(tickets)-1].Id
	return tickets, &nextCursor, nil
}

func (r *TicketRepository) UpdateTickets(ctx context.Context, tickets []domain.Ticket) error {

	if len(tickets) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, 0, len(tickets))

	for _, ticket := range tickets {
		filter := bson.M{"_id": ticket.Id}
		update := bson.M{
			"$set": bson.M{
				"state": ticket.State,
			},
		}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}

	_, err := r.collection.BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("error bulk updating tickets: %w", err)
	}

	return nil
}
