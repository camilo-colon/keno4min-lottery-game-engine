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
	ID        string      `bson:"_id,omitempty" json:"id"`
	Round     int64       `bson:"round" json:"round"`
	Cupon     string      `bson:"cupon" json:"cupon"`
	State     TicketState `bson:"state" json:"state"`
	Win       int64       `bson:"win" json:"win"`
	// Jackpot es el premio de jackpot ganado por este ticket (0 si no ganó pozo).
	Jackpot int64 `bson:"jackpot" json:"jackpot"`
	Bets      []Bet       `bson:"bets" json:"bets"`
	Total     int64       `bson:"total" json:"total"`
	GameID    string      `bson:"game_id" json:"gameId"`
	Balls     []int64     `bson:"balls" json:"balls"`
	CreatedBy string      `bson:"created_by" json:"createdBy"`
	ClubID    string      `bson:"club_id" json:"clubId"`
	AdminID   string      `bson:"admin_id" json:"adminId"`
	DatePay   *time.Time  `bson:"date_pay,omitempty" json:"datePay,omitempty"`
	CreatedAt time.Time   `bson:"created_at" json:"createdAt"`
	Version   int64       `bson:"version" json:"version"`
}

// Bet representa una apuesta individual dentro de un ticket.
type Bet struct {
	Nums    []int64 `bson:"nums" json:"nums"`
	Money   int64   `bson:"money" json:"money"`
	Bitmask Bitmask `bson:"bitmask" json:"bitmask"`
}

// Payout calcula lo que paga esta apuesta contra las balotas sorteadas: el
// factor de PaymentTable[picks][aciertos] aplicado al monto apostado.
func (b Bet) Payout(balls Bitmask) int64 {
	picks := b.Bitmask.Count()
	hits := b.Bitmask.Matches(balls)

	factors, ok := PaymentTable[picks]
	if !ok || hits >= len(factors) {
		return 0
	}
	return b.Money * factors[hits] / 100
}

// Payout es lo que gana el jugador con este ticket: la suma de sus apuestas.
func (t Ticket) Payout(balls Bitmask) int64 {
	var total int64
	for _, bet := range t.Bets {
		total += bet.Payout(balls)
	}
	return total
}

// HouseProfit es la utilidad de la casa por este ticket: lo apostado menos lo
// pagado al jugador. Puede ser negativa (el jugador ganó más de lo que apostó).
func (t Ticket) HouseProfit(balls Bitmask) int64 {
	return t.Total - t.Payout(balls)
}
