package adapters

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClubStore struct {
	collection *mongo.Collection
}

func NewClubStore(db *mongo.Database) *ClubStore {
	return &ClubStore{
		collection: db.Collection("clubs"),
	}
}

// FindByID busca un club por su ID.
func (r *ClubStore) FindByID(ctx context.Context, clubID string) (*domain.Club, error) {
	var club domain.Club
	if err := r.collection.FindOne(ctx, bson.M{"_id": clubID}).Decode(&club); err != nil {
		return nil, fmt.Errorf("club %s not found: %w", clubID, err)
	}
	return &club, nil
}

// IncrementJackpot suma de forma atómica un monto al pozo jp1 del club y
// devuelve el estado del jackpot ya incrementado.
func (r *ClubStore) IncrementJackpot(ctx context.Context, clubID string, amount int64) (*domain.Jackpot, error) {
	filter := bson.M{"_id": clubID}
	update := bson.M{"$inc": bson.M{"jp1.value": amount}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var club domain.Club
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&club); err != nil {
		return nil, fmt.Errorf("error incrementing jackpot for club %s: %w", clubID, err)
	}
	return &club.JP1, nil
}

// ResetJackpot reemplaza el pozo jp1 del club con un jackpot fresco.
func (r *ClubStore) ResetJackpot(ctx context.Context, clubID string, jackpot domain.Jackpot) error {
	filter := bson.M{"_id": clubID}
	update := bson.M{"$set": bson.M{"jp1": jackpot}}

	if _, err := r.collection.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("error resetting jackpot for club %s: %w", clubID, err)
	}
	return nil
}
