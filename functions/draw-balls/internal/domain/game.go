package domain

import "time"

type GameStatus string

var (
	BETTING GameStatus = "BETTING"
	DRAWN   GameStatus = "DRAWN"
)

type GameBalls struct {
	Nums []uint64 `bson:"nums" json:"nums"`
	Mask BitMask  `bson:"mask" json:"mask"`
}

type Game struct {
	ID        string     `bson:"_id,omitempty" json:"id"`
	Round     uint64     `bson:"round" json:"round"`
	Status    GameStatus `bson:"status" json:"status"`
	Idv       string     `bson:"idv,omitempty" json:"idv,omitempty"`
	Balls     *GameBalls `bson:"balls,omitempty" json:"balls,omitempty"`
	StartsAt  time.Time  `bson:"starts_at" json:"startsAt"`
	ClosesAt  time.Time  `bson:"closes_at" json:"closesAt"`
	CreatedAt time.Time  `bson:"created_at" json:"createdAt"`
}

func (g *Game) DrawBalls(idv string, balls GameBalls) {
	g.Idv = idv
	g.Balls = &balls
	g.Status = DRAWN
}
