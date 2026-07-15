package domain_test

import (
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/domain"
)

// realTicket reproduce el ticket de ejemplo: 3 apuestas de 50.000, total 150.000.
// Solo la primera apuesta {33,54,24} acierta las 3 balotas.
func realTicket() domain.Ticket {
	return domain.Ticket{
		Total: 150000,
		Bets: []domain.Bet{
			{Money: 50000, Bitmask: mask(33, 54, 24)},
			{Money: 50000, Bitmask: mask(57, 22, 34)},
			{Money: 50000, Bitmask: mask(38, 52, 4)},
		},
	}
}

func TestTicketPayout(t *testing.T) {
	tests := []struct {
		name   string
		ticket domain.Ticket
		balls  domain.Bitmask
		want   int64
	}{
		{
			name:   "ticket real: solo una apuesta acierta",
			ticket: realTicket(),
			balls:  gameBalls,
			want:   2_500_000, // coincide con el win almacenado del ticket
		},
		{
			name: "todas las apuestas fallan no pagan",
			ticket: domain.Ticket{
				Total: 150000,
				Bets: []domain.Bet{
					{Money: 50000, Bitmask: mask(1, 2, 3)},
					{Money: 50000, Bitmask: mask(5, 6, 8)},
				},
			},
			balls: gameBalls,
			want:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ticket.Payout(tt.balls); got != tt.want {
				t.Errorf("Payout() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTicketResolveWinning(t *testing.T) {
	ticket := realTicket()
	ticket.State = domain.DRAWING

	balls := domain.GameBalls{Nums: []int64{7, 9, 11, 13, 16, 18, 24, 25, 26, 30, 33, 42, 44, 54, 55, 58, 62, 66, 77, 80}, Mask: gameBalls}
	ticket.Resolve(balls)

	if ticket.Win != 2_500_000 {
		t.Errorf("Win = %d, want 2500000", ticket.Win)
	}
	if ticket.State != domain.WINNING {
		t.Errorf("State = %q, want WINNING", ticket.State)
	}
	if len(ticket.Balls) != len(balls.Nums) {
		t.Fatalf("Balls no se copiaron: got %v", ticket.Balls)
	}
	for i, n := range balls.Nums {
		if ticket.Balls[i] != n {
			t.Errorf("Balls[%d] = %d, want %d", i, ticket.Balls[i], n)
		}
	}
}

func TestTicketResolveLosing(t *testing.T) {
	ticket := domain.Ticket{
		State: domain.DRAWING,
		Total: 150000,
		Bets: []domain.Bet{
			{Money: 50000, Bitmask: mask(1, 2, 3)},
		},
	}
	balls := domain.GameBalls{Nums: []int64{7, 9, 11}, Mask: gameBalls}

	ticket.Resolve(balls)

	if ticket.Win != 0 {
		t.Errorf("Win = %d, want 0", ticket.Win)
	}
	if ticket.State != domain.LOSS {
		t.Errorf("State = %q, want LOSS", ticket.State)
	}
	if len(ticket.Balls) != 3 {
		t.Errorf("Balls no se copiaron: got %v", ticket.Balls)
	}
}
