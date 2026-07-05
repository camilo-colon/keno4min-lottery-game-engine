package domain

import (
	"time"

	"github.com/oklog/ulid/v2"
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

func NewGame(round uint64) *Game {
	now := time.Now()
	return &Game{
		ID:        ulid.Make().String(),
		Status:    BETTING,
		Round:     round,
		StartsAt:  now,
		ClosesAt:  now.Add(157 * time.Second),
		CreatedAt: now,
	}
}
