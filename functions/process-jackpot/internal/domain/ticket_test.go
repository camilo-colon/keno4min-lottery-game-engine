package domain_test

import (
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
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

func TestTicketHouseProfit(t *testing.T) {
	tests := []struct {
		name   string
		ticket domain.Ticket
		balls  domain.Bitmask
		want   int64
	}{
		{
			name:   "ticket real: la casa pierde (pagó más de lo apostado)",
			ticket: realTicket(),
			balls:  gameBalls,
			want:   150000 - 2_500_000, // -2.350.000
		},
		{
			name: "todas fallan: la casa gana todo lo apostado",
			ticket: domain.Ticket{
				Total: 150000,
				Bets: []domain.Bet{
					{Money: 50000, Bitmask: mask(1, 2, 3)},
				},
			},
			balls: gameBalls,
			want:  150000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ticket.HouseProfit(tt.balls); got != tt.want {
				t.Errorf("HouseProfit() = %d, want %d", got, tt.want)
			}
		})
	}
}
