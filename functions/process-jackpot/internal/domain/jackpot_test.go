package domain_test

import (
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

func TestJackpotIncrementFor(t *testing.T) {
	tests := []struct {
		name   string
		jp     domain.Jackpot
		profit int64
		want   int64
	}{
		{
			name:   "utilidad positiva aplica el porcentaje del jp",
			jp:     domain.Jackpot{Percent: 1},
			profit: 8000, // ejemplo: apostó 10000, pagó 2000
			want:   80,   // 8000 * 1 / 100
		},
		{
			name:   "utilidad negativa no aporta",
			jp:     domain.Jackpot{Percent: 1},
			profit: -24500, // ejemplo: apostó 500, pagó 25000
			want:   0,
		},
		{
			name:   "utilidad cero no aporta",
			jp:     domain.Jackpot{Percent: 1},
			profit: 0,
			want:   0,
		},
		{
			name:   "porcentaje 3 sobre utilidad grande",
			jp:     domain.Jackpot{Percent: 3},
			profit: 2_000_000,
			want:   60_000,
		},
		{
			name:   "trunca al aplicar el porcentaje",
			jp:     domain.Jackpot{Percent: 1},
			profit: 150,
			want:   1, // 1.5 → 1
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.jp.IncrementFor(tt.profit); got != tt.want {
				t.Errorf("IncrementFor() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestJackpotShouldPlay(t *testing.T) {
	tests := []struct {
		name string
		jp   domain.Jackpot
		want bool
	}{
		{"value bajo el target no juega", domain.Jackpot{Value: 16_000_000, Target: 23_281_912}, false},
		{"value igual al target juega", domain.Jackpot{Value: 23_281_912, Target: 23_281_912}, true},
		{"value sobre el target juega", domain.Jackpot{Value: 25_000_000, Target: 23_281_912}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.jp.ShouldPlay(); got != tt.want {
				t.Errorf("ShouldPlay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewJackpotFromConfig(t *testing.T) {
	cfg := domain.JackpotConfig{
		Min:     15_000_000,
		Max:     25_000_000,
		Percent: 3,
		Seed:    5_000_000,
	}

	jp := domain.NewJackpotFromConfig(cfg, 20_000_000)

	if jp.Value != cfg.Seed {
		t.Errorf("Value = %d, want seed %d", jp.Value, cfg.Seed)
	}
	if jp.Target != 20_000_000 {
		t.Errorf("Target = %d, want 20000000", jp.Target)
	}
	if jp.Percent != cfg.Percent {
		t.Errorf("Percent = %d, want config %d", jp.Percent, cfg.Percent)
	}
	if jp.Min != cfg.Min || jp.Max != cfg.Max {
		t.Errorf("Min/Max = %d/%d, want %d/%d", jp.Min, jp.Max, cfg.Min, cfg.Max)
	}
}
