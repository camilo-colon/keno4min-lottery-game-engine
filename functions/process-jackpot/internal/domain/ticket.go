package domain

import "time"

// TicketState representa el estado de un ticket.
type TicketState string

var (
	LOSS     TicketState = "LOSS"
	PAYED    TicketState = "PAYED"
	WINNING  TicketState = "WINNING"
	CANCELED TicketState = "CANCELED"
)

// Ticket representa una jugada de un cliente en un juego.
type Ticket struct {
	ID     string      `bson:"_id,omitempty" json:"id"`
	Round  int64       `bson:"round" json:"round"`
	Coupon string      `bson:"coupon" json:"coupon"`
	State  TicketState `bson:"state" json:"state"`
	// Win es el payout ya calculado y persistido por la Lambda update-tickets
	// (corre entre DrawBalls y ProcessJackpot).
	Win int64 `bson:"win" json:"win"`
	// Jackpot es el premio de jackpot ganado por este ticket (0 si no ganó pozo).
	Jackpot   int64      `bson:"jackpot" json:"jackpot"`
	Bets      []Bet      `bson:"bets" json:"bets"`
	Total     int64      `bson:"total" json:"total"`
	GameID    string     `bson:"game_id" json:"gameId"`
	Balls     []int64    `bson:"balls" json:"balls"`
	CreatedBy string     `bson:"created_by" json:"createdBy"`
	ClubID    string     `bson:"club_id" json:"clubId"`
	AdminID   string     `bson:"admin_id" json:"adminId"`
	DatePay   *time.Time `bson:"date_pay,omitempty" json:"datePay,omitempty"`
	CreatedAt time.Time  `bson:"created_at" json:"createdAt"`
	Version   int64      `bson:"version" json:"version"`
}

// Bet representa una apuesta individual dentro de un ticket.
type Bet struct {
	Nums    []int64 `bson:"nums" json:"nums"`
	Money   int64   `bson:"money" json:"money"`
	Bitmask Bitmask `bson:"bitmask" json:"bitmask"`
}

// HouseProfit es la utilidad de la casa por este ticket: lo apostado menos lo
// pagado al jugador (Win, ya calculado y persistido por update-tickets). Puede
// ser negativa (el jugador ganó más de lo que apostó).
func (t Ticket) HouseProfit() int64 {
	return t.Total - t.Win
}

// AwardJackpot le asigna al ticket el pozo ganado y lo deja cobrable.
//
// Marcar WINNING es parte del premio, no un detalle: update-tickets deriva el
// estado ÚNICAMENTE de Win (WINNING si win > 0, LOSS si no) y corre ANTES del
// sorteo del pozo, cuando el jackpot todavía no existe. Un ticket sin aciertos
// que gana el pozo llega acá en LOSS; si solo escribiéramos el monto quedaría
// LOSS con jackpot asignado: un premio impagable, porque la caja solo cobra
// tickets WINNING.
func (t *Ticket) AwardJackpot(amount int64) {
	t.Jackpot = amount
	t.State = WINNING
}
