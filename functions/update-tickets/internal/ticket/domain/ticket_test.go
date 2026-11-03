package domain

import (
	"slices"
	"testing"
	"time"

	"github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/events"
)

var (
	balls     = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	settledAt = time.Date(2026, 11, 2, 21, 30, 0, 0, time.UTC)
)

func TestBetPrize(t *testing.T) {
	tests := []struct {
		name string
		bet  Bet
		want int64
	}{
		{"1 spot 1 acierto paga x3.5", Bet{Nums: []int{1}, ValueCents: 1000}, 3500},
		{"1 spot 0 aciertos no paga", Bet{Nums: []int{80}, ValueCents: 1000}, 0},
		{"2 spots 1 acierto paga x1", Bet{Nums: []int{1, 80}, ValueCents: 1000}, 1000},
		{"2 spots 2 aciertos paga x10", Bet{Nums: []int{1, 2}, ValueCents: 1000}, 10000},
		{"4 spots 1 acierto no paga", Bet{Nums: []int{1, 78, 79, 80}, ValueCents: 1000}, 0},
		{"10 spots 0 aciertos paga x2", Bet{Nums: []int{71, 72, 73, 74, 75, 76, 77, 78, 79, 80}, ValueCents: 1000}, 2000},
		{"10 spots 10 aciertos paga x10000", Bet{Nums: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, ValueCents: 1000}, 10000000},
		{"redondea hacia abajo", Bet{Nums: []int{1}, ValueCents: 1}, 3},
		{"spots fuera de la tabla no pagan", Bet{Nums: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, ValueCents: 1000}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bet.Prize(balls); got != tt.want {
				t.Fatalf("Prize() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSettleWinningSumsAllBets(t *testing.T) {
	ticket := Ticket{Status: StatusPending, Bets: []Bet{
		{Nums: []int{1}, ValueCents: 1000},   // 3500
		{Nums: []int{80}, ValueCents: 1000},  // 0
		{Nums: []int{1, 2}, ValueCents: 500}, // 5000
	}}
	ticket.Settle(balls, settledAt)
	if ticket.Status != StatusWinning || ticket.PrizeCents != 8500 {
		t.Fatalf("got status %s prize %d", ticket.Status, ticket.PrizeCents)
	}
}

func TestSettleLost(t *testing.T) {
	ticket := Ticket{Status: StatusPending, PrizeCents: 999, Bets: []Bet{{Nums: []int{80}, ValueCents: 1000}}}
	ticket.Settle(balls, settledAt)
	if ticket.Status != StatusLost || ticket.PrizeCents != 0 {
		t.Fatalf("got status %s prize %d", ticket.Status, ticket.PrizeCents)
	}
}

func TestSettleKeepsACopyOfTheBalls(t *testing.T) {
	drawn := slices.Clone(balls)
	ticket := Ticket{Status: StatusPending, Bets: []Bet{{Nums: []int{80}, ValueCents: 1000}}}
	ticket.Settle(drawn, settledAt)
	if !slices.Equal(ticket.Balls, balls) {
		t.Fatalf("got balls %v", ticket.Balls)
	}
	drawn[0] = 99
	if ticket.Balls[0] != 1 {
		t.Fatal("ticket balls alias the draw balls")
	}
}

func TestSettleRecordsTicketSettled(t *testing.T) {
	ticket := Ticket{ID: "ticket-1", Status: StatusPending, Bets: []Bet{{Nums: []int{1}, ValueCents: 1000}}}
	ticket.Settle(balls, settledAt)

	recorded := ticket.PullDomainEvents()
	if len(recorded) != 1 {
		t.Fatalf("recorded %d events, want 1", len(recorded))
	}
	got, ok := recorded[0].(events.TicketSettled)
	if !ok {
		t.Fatalf("got %T, want events.TicketSettled", recorded[0])
	}
	if got.TicketID != "ticket-1" || got.Game != "keno4min" || got.Status != "winning" || got.PrizeCents != 3500 {
		t.Fatalf("unexpected event: %+v", got)
	}
	if !got.OccurredOnUtc().Equal(settledAt) || got.Type() != "cronos.keno4min.ticket.settled" {
		t.Fatalf("unexpected event metadata: %+v", got)
	}
}
