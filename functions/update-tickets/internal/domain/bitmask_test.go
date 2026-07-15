package domain_test

import (
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/domain"
)

func TestBitmaskCount(t *testing.T) {
	tests := []struct {
		name string
		mask domain.Bitmask
		want int
	}{
		{"máscara vacía", mask(), 0},
		{"tres números en mask1", mask(1, 2, 3), 3},
		{"números altos en mask2", mask(65, 80), 2},
		{"a ambos lados del bit 64", mask(64, 65), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mask.Count(); got != tt.want {
				t.Errorf("Count() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBitmaskMatches(t *testing.T) {
	tests := []struct {
		name string
		a    domain.Bitmask
		b    domain.Bitmask
		want int
	}{
		{"sin coincidencias", mask(1, 2, 3), mask(4, 5, 6), 0},
		{"coincidencia parcial", mask(1, 2, 3), mask(2, 3, 4), 2},
		{"coincidencia total", mask(7, 9, 11), mask(7, 9, 11), 3},
		{"coincidencia en mask2", mask(65, 70, 80), mask(70, 80), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Matches(tt.b); got != tt.want {
				t.Errorf("Matches() = %d, want %d", got, tt.want)
			}
		})
	}
}
