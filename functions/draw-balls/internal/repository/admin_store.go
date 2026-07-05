package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AccumulatedTotals struct {
	AdminId     string  `bson:"admin_id" json:"adminId"`
	TotalIncome float64 `bson:"total_income" json:"totalIncome"`
	TotalPaid   float64 `bson:"total_paid" json:"totalPaid"`
}

type AdminStore struct {
	db *mongo.Database
}

func NewAdminStore(db *mongo.Database) *AdminStore {
	return &AdminStore{
		db: db,
	}
}

func (s *AdminStore) GetAccumulatedTotalsByAdmin(ctx context.Context, adminID string) (*AccumulatedTotals, error) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -30).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"admin_id": adminID,
				"date": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":          nil,
				"total_income": bson.M{"$sum": "$total_income"},
				"total_paid":   bson.M{"$sum": "$total_paid"},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":          0,
				"total_income": 1,
				"total_paid":   1,
			},
		},
	}

	cursor, err := s.db.Collection("admin_stats").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result AccumulatedTotals
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	}
	result.AdminId = adminID
	return &result, nil
}
