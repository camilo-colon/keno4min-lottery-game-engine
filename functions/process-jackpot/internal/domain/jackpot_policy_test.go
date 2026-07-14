package domain_test

import (
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

// losingTicket construye un ticket cuyas apuestas no aciertan (payout 0), por lo
// que su utilidad para la casa es igual a lo apostado.
func losingTicket(total int64) domain.Ticket {
	return domain.Ticket{
		Total: total,
		Bets:  []domain.Bet{{Money: total, Bitmask: mask(1, 2, 3)}},
	}
}

// bigWinTicket construye un ticket que hace perder a la casa: acierta 3 de 3.
func bigWinTicket(total int64) domain.Ticket {
	return domain.Ticket{
		Total: total,
		Bets:  []domain.Bet{{Money: 50000, Bitmask: mask(33, 54, 24)}},
	}
}

func TestJackpotIncrement(t *testing.T) {
	jp := domain.Jackpot{Percent: 1}

	tests := []struct {
		name    string
		jp      domain.Jackpot
		tickets []domain.Ticket
		want    int64
	}{
		{
			name: "solo aportan los tickets con utilidad positiva",
			jp:   jp,
			tickets: []domain.Ticket{
				losingTicket(150000), // utilidad +150000
				bigWinTicket(50000),  // pagó 2.500.000 → utilidad negativa, excluido
			},
			want: 1500, // 150000 * 1 / 100
		},
		{
			name:    "club con pérdida neta no aporta",
			jp:      jp,
			tickets: []domain.Ticket{bigWinTicket(50000)},
			want:    0,
		},
		{
			name:    "sin tickets no aporta",
			jp:      jp,
			tickets: nil,
			want:    0,
		},
		{
			name: "suma antes de aplicar el porcentaje (una sola división)",
			jp:   domain.Jackpot{Percent: 1},
			tickets: []domain.Ticket{
				losingTicket(150), // +150
				losingTicket(150), // +150
			},
			want: 3, // (150+150)*1/100 = 3, no 1+1 por ticket
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.JackpotIncrement(tt.jp, tt.tickets, gameBalls); got != tt.want {
				t.Errorf("JackpotIncrement() = %d, want %d", got, tt.want)
			}
		})
	}
}

