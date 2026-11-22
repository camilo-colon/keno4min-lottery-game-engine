package domain

import (
	"fmt"

	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
)

type Bet struct {
	Numbers    bitmask.Mask
	StakeCents int64
}

func New(numbers bitmask.Mask, stakeCents int64) (Bet, error) {
	spots := numbers.Count()
	if spots < 1 || spots > 10 {
		return Bet{}, fmt.Errorf("bet has %d numbers, want 1 to 10", spots)
	}
	if stakeCents <= 0 {
		return Bet{}, fmt.Errorf("bet stake is %d cents, want more than 0", stakeCents)
	}
	return Bet{Numbers: numbers, StakeCents: stakeCents}, nil
}

func (b Bet) PrizeCents(drawn bitmask.Mask) int64 {
	spots := b.Numbers.Count()
	hits := bitmask.Hits(b.Numbers, drawn)
	return b.StakeCents * paymentTable[spots][hits] / 100
}
