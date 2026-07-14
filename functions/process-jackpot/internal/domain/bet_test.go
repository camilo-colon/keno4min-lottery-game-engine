package domain_test

import (
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

func TestBetPayout(t *testing.T) {
	tests := []struct {
		name  string
		bet   domain.Bet
		balls domain.Bitmask
		want  int64
	}{
		{
			name:  "3 números, 3 aciertos paga factor 5000",
			bet:   domain.Bet{Money: 50000, Bitmask: mask(33, 54, 24)},
			balls: mask(33, 54, 24, 7, 9),
			want:  2_500_000, // 50000 * 5000 / 100
		},
		{
			name:  "3 números, 2 aciertos paga factor 200",
			bet:   domain.Bet{Money: 50000, Bitmask: mask(33, 54, 24)},
			balls: mask(33, 54, 7, 9),
			want:  100_000, // 50000 * 200 / 100
		},
		{
			name:  "3 números, 1 acierto no paga",
			bet:   domain.Bet{Money: 50000, Bitmask: mask(33, 54, 24)},
			balls: mask(33, 7, 9),
			want:  0,
		},
		{
			name:  "1 número, 1 acierto paga factor 350",
			bet:   domain.Bet{Money: 10000, Bitmask: mask(7)},
			balls: mask(7, 9, 11),
			want:  35_000, // 10000 * 350 / 100
		},
		{
			name:  "2 números, 1 acierto paga factor 10",
			bet:   domain.Bet{Money: 50000, Bitmask: mask(7, 9)},
			balls: mask(7, 40),
			want:  5_000, // 50000 * 10 / 100
		},
		{
			name:  "2 números, 2 aciertos paga factor 100",
			bet:   domain.Bet{Money: 50000, Bitmask: mask(7, 9)},
			balls: mask(7, 9),
			want:  50_000, // 50000 * 100 / 100
		},
		{
			name:  "trunca los centavos del pago",
			bet:   domain.Bet{Money: 333, Bitmask: mask(7)},
			balls: mask(7),
			want:  1_165, // 333 * 350 / 100 = 1165.5 → 1165
		},
		{
			name:  "apuesta vacía no paga",
			bet:   domain.Bet{Money: 50000, Bitmask: mask()},
			balls: mask(7, 9),
			want:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bet.Payout(tt.balls); got != tt.want {
				t.Errorf("Payout() = %d, want %d", got, tt.want)
			}
		})
	}
}
