package domain

import "time"

// GameStatus representa el estado de un juego.
type GameStatus string

var (
	BETTING GameStatus = "BETTING"
	DRAWN   GameStatus = "DRAWN"
)

// Game representa una ronda de juego.
type Game struct {
	ID        string     `bson:"_id,omitempty" json:"id"`
	Round     int64      `bson:"round" json:"round"`
	Status    GameStatus `bson:"status" json:"status"`
	StartsAt  time.Time  `bson:"starts_at" json:"startsAt"`
	ClosesAt  time.Time  `bson:"closes_at" json:"closesAt"`
	CreatedAt time.Time  `bson:"created_at" json:"createdAt"`
	// Balls se llena al ejecutar el sorteo; nil mientras el juego está en BETTING.
	Balls *GameBalls `bson:"balls,omitempty" json:"balls,omitempty"`
	Idv   string     `bson:"idv,omitempty" json:"idv,omitempty"`
}

// GameBalls contiene las balotas sorteadas y su máscara de bits.
type GameBalls struct {
	Nums []int64 `bson:"nums" json:"nums"`
	Mask Bitmask `bson:"mask" json:"mask"`
}
