package domain_test

import "github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/domain"

// mask construye un Bitmask a partir de números (1..80), igual que en producción:
// bit (n-1), mask1 para 1..64 y mask2 para 65..80.
func mask(nums ...int64) domain.Bitmask {
	var b domain.Bitmask
	for _, n := range nums {
		pos := n - 1
		if pos < 64 {
			b.Mask1 |= int64(1) << uint(pos)
		} else {
			b.Mask2 |= int64(1) << uint(pos-64)
		}
	}
	return b
}

// gameBalls son las balotas reales del sorteo de ejemplo (game DRAWN).
var gameBalls = mask(7, 9, 11, 13, 16, 18, 24, 25, 26, 30, 33, 42, 44, 54, 55, 58, 62, 66, 77, 80)
