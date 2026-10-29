package domain

import "slices"

type Status string

const (
	StatusPending   Status = "pending"
	StatusCancelled Status = "cancelled"
	StatusLost      Status = "lost"
	StatusWinning   Status = "winning"
)

type Bet struct {
	Nums       []int
	ValueCents int64
}

type Ticket struct {
	ID         string
	TIR        uint64
	Bets       []Bet
	Status     Status
	PrizeCents int64
	Balls      []int
}

// Prize devuelve el pago de la apuesta en centavos según las balotas del sorteo.
func (b Bet) Prize(balls []int) int64 {
	hits := 0
	for _, num := range b.Nums {
		if slices.Contains(balls, num) {
			hits++
		}
	}
	factors, ok := PaymentTable[len(b.Nums)]
	if !ok || hits >= len(factors) {
		return 0
	}
	return b.ValueCents * factors[hits] / 100
}

// Settle suma el premio de todas las apuestas, marca el ticket como ganador o perdedor
// y conserva las balotas con las que se liquidó.
func (t *Ticket) Settle(balls []int) {
	t.Balls = slices.Clone(balls)
	t.PrizeCents = 0
	for _, bet := range t.Bets {
		t.PrizeCents += bet.Prize(balls)
	}
	t.Status = StatusLost
	if t.PrizeCents > 0 {
		t.Status = StatusWinning
	}
}
