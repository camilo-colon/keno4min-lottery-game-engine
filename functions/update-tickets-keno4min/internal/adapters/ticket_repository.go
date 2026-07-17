package adapters

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TicketStore struct {
	collection *mongo.Collection
}

func NewTicketStore(db *mongo.Database) *TicketStore {
	return &TicketStore{
		collection: db.Collection("tickets"),
	}
}

// FindPendingByGame devuelve una página de tickets PENDING del juego, ordenados
// por _id ascendente. Ver ports.TicketRepository para la regla de por qué el
// filtro es PENDING y no "!= CANCELED".
func (r *TicketStore) FindPendingByGame(ctx context.Context, gameID string, cursor *string, limit int64) ([]domain.Ticket, *string, error) {
	if limit <= 0 {
		limit = 100
	}

	filter := bson.M{
		"game_id": gameID,
		"state":   domain.PENDING,
	}
	if cursor != nil && *cursor != "" {
		filter["_id"] = bson.M{"$gt": *cursor}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(limit)

	mongoCursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("error finding drawing tickets for game %s: %w", gameID, err)
	}
	defer mongoCursor.Close(ctx)

	var tickets []domain.Ticket
	if err := mongoCursor.All(ctx, &tickets); err != nil {
		return nil, nil, fmt.Errorf("error decoding tickets for game %s: %w", gameID, err)
	}

	if len(tickets) == 0 {
		return tickets, nil, nil
	}

	nextCursor := tickets[len(tickets)-1].ID
	return tickets, &nextCursor, nil
}

// UpdateTickets persiste win, state y balls de los tickets recibidos en un
// solo BulkWrite.
func (r *TicketStore) UpdateTickets(ctx context.Context, tickets []domain.Ticket) error {
	if len(tickets) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, 0, len(tickets))
	for _, ticket := range tickets {
		filter := bson.M{"_id": ticket.ID}
		update := bson.M{
			"$set": bson.M{
				"win":   ticket.Win,
				"state": ticket.State,
				"balls": ticket.Balls,
			},
		}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}

	if _, err := r.collection.BulkWrite(ctx, models); err != nil {
		return fmt.Errorf("error bulk updating tickets: %w", err)
	}
	return nil
}
