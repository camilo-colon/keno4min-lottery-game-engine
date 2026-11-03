package domain

import (
	"slices"
	"time"

	"github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/common"
	"github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/events"
)

// Game identifies this product in the events consumed by other systems.
const Game = "keno4min"

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
	common.AggregateRoot

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

// Settle suma el premio de todas las apuestas, marca el ticket como ganador o perdedor,
// conserva las balotas con las que se liquidó y registra el evento TicketSettled.
func (t *Ticket) Settle(balls []int, settledAt time.Time) {
	t.Balls = slices.Clone(balls)
	t.PrizeCents = 0
	for _, bet := range t.Bets {
		t.PrizeCents += bet.Prize(balls)
	}
	t.Status = StatusLost
	if t.PrizeCents > 0 {
		t.Status = StatusWinning
	}
	t.Record(events.TicketSettled{
		BaseEvent:  common.NewBaseEvent(settledAt),
		TicketID:   t.ID,
		Game:       Game,
		Status:     string(t.Status),
		PrizeCents: t.PrizeCents,
	})
}
