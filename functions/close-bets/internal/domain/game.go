package domain

import (
	"context"
	"time"
)

type GameStatus string

var (
	BETTING GameStatus = "BETTING"
	DRAWN   GameStatus = "DRAWN"
)

type Game struct {
	ID        string     `bson:"_id,omitempty" json:"id"`
	Round     uint64     `bson:"round" json:"round"`
	Status    GameStatus `bson:"status" json:"status"`
	StartsAt  time.Time  `bson:"starts_at" json:"startsAt"`
	ClosesAt  time.Time  `bson:"closes_at" json:"closesAt"`
	CreatedAt time.Time  `bson:"created_at" json:"createdAt"`
}

type GameRepository interface {
	UpdateStatus(ctx context.Context, gameID string, from, to GameStatus) error
}
