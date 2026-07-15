package domain_test

import (
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

func TestTicketHouseProfit(t *testing.T) {
	tests := []struct {
		name   string
		ticket domain.Ticket
		want   int64
	}{
		{
			name:   "el ticket gana más de lo apostado: la casa pierde",
			ticket: domain.Ticket{Total: 150000, Win: 2_500_000},
			want:   150000 - 2_500_000, // -2.350.000
		},
		{
			name:   "el ticket no gana nada: la casa se queda todo lo apostado",
			ticket: domain.Ticket{Total: 150000, Win: 0},
			want:   150000,
		},
		{
			name:   "el ticket gana menos de lo apostado: la casa retiene la diferencia",
			ticket: domain.Ticket{Total: 1000, Win: 400},
			want:   600,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ticket.HouseProfit(); got != tt.want {
				t.Errorf("HouseProfit() = %d, want %d", got, tt.want)
			}
		})
	}
}
