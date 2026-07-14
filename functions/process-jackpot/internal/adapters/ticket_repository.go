package adapters

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
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

// FindClubIDsByGame devuelve los club_id distintos entre los tickets NO cancelados
// del juego. Un club con solo tickets cancelados no participa del jackpot.
func (r *TicketStore) FindClubIDsByGame(ctx context.Context, gameID string) ([]string, error) {
	filter := bson.M{"game_id": gameID, "state": bson.M{"$ne": domain.CANCELED}}
	values, err := r.collection.Distinct(ctx, "club_id", filter)
	if err != nil {
		return nil, fmt.Errorf("error finding club ids for game %s: %w", gameID, err)
	}

	clubIDs := make([]string, 0, len(values))
	for _, v := range values {
		if id, ok := v.(string); ok {
			clubIDs = append(clubIDs, id)
		}
	}
	return clubIDs, nil
}

// FindByClubAndGame devuelve los tickets NO cancelados de un club en un juego.
func (r *TicketStore) FindByClubAndGame(ctx context.Context, clubID, gameID string) ([]domain.Ticket, error) {
	filter := bson.M{"game_id": gameID, "club_id": clubID, "state": bson.M{"$ne": domain.CANCELED}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error finding tickets for club %s in game %s: %w", clubID, gameID, err)
	}
	defer cursor.Close(ctx)

	var tickets []domain.Ticket
	if err := cursor.All(ctx, &tickets); err != nil {
		return nil, fmt.Errorf("error decoding tickets: %w", err)
	}
	return tickets, nil
}

// AssignJackpot asigna el premio del pozo al ticket ganador.
func (r *TicketStore) AssignJackpot(ctx context.Context, ticketID string, amount int64) error {
	filter := bson.M{"_id": ticketID}
	update := bson.M{"$set": bson.M{"jackpot": amount}}

	if _, err := r.collection.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("error assigning jackpot to ticket %s: %w", ticketID, err)
	}
	return nil
}
