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

func TestJackpotIncrement(t *testing.T) {
	jp := domain.Jackpot{Percent: 1}

	tests := []struct {
		name    string
		jp      domain.Jackpot
		tickets []domain.Ticket
		want    int64
	}{
		{
			name: "las perdidas restan del neto (neto positivo incrementa)",
			jp:   jp,
			tickets: []domain.Ticket{
				losingTicket(2000), // apostado 2000, gana 0 -> +2000
				// apostado 500, acierta 2 de 3 (factor 200) -> gana 1000 -> -500
				{Total: 500, Bets: []domain.Bet{{Money: 500, Bitmask: mask(7, 9, 1)}}},
			},
			want: 15, // neto 1500 * 1 / 100
		},
		{
			name: "neto negativo no incrementa",
			jp:   jp,
			tickets: []domain.Ticket{
				// apostado 500, acierta 3 de 3 (factor 5000) -> gana 25000 -> -24500
				{Total: 500, Bets: []domain.Bet{{Money: 500, Bitmask: mask(7, 9, 11)}}},
				losingTicket(2000), // apostado 2000, gana 0 -> +2000
			},
			want: 0, // neto -22500
		},
		{
			name:    "sin tickets no aporta",
			jp:      jp,
			tickets: nil,
			want:    0,
		},
		{
			name: "suma el neto antes de aplicar el porcentaje (una sola division)",
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
