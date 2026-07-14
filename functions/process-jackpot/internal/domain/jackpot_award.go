package domain

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// JackpotAward es el registro histórico e inmutable de un jackpot ganado: qué
// ticket lo ganó, en qué club/juego, por cuánto y cuándo. A diferencia del pozo
// vivo (Jackpot, value object), un premio SÍ es una entidad: tiene identidad.
type JackpotAward struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	ClubID    string    `bson:"club_id" json:"clubId"`
	GameID    string    `bson:"game_id" json:"gameId"`
	TicketID  string    `bson:"ticket_id" json:"ticketId"`
	Cupon     string    `bson:"cupon" json:"cupon"`
	Round     int64     `bson:"round" json:"round"`
	Value     int64     `bson:"value" json:"value"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
}

// NewJackpotAward construye el registro de un premio a partir del ticket ganador
// y el monto del pozo pagado.
func NewJackpotAward(winner Ticket, value int64) JackpotAward {
	return JackpotAward{
		ID:        ulid.Make().String(),
		ClubID:    winner.ClubID,
		GameID:    winner.GameID,
		TicketID:  winner.ID,
		Cupon:     winner.Cupon,
		Round:     winner.Round,
		Value:     value,
		CreatedAt: time.Now().UTC(),
	}
}
