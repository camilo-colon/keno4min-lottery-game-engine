package domain

import "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"

type AdminBets struct {
	AdminNodeID string
	Bets        []Bet
}

func (a AdminBets) StakeCents() int64 {
	var stakeCents int64
	for _, bet := range a.Bets {
		stakeCents += bet.StakeCents
	}
	return stakeCents
}

func (a AdminBets) PrizeCents(drawn bitmask.Mask) int64 {
	var prizeCents int64
	for _, bet := range a.Bets {
		prizeCents += bet.PrizeCents(drawn)
	}
	return prizeCents
}

func (a AdminBets) PayoutRates(videos []bitmask.Mask) []float64 {
	stakeCents := float64(a.StakeCents())
	payoutRates := make([]float64, len(videos))
	for i, drawn := range videos {
		payoutRates[i] = float64(a.PrizeCents(drawn)) / stakeCents
	}
	return payoutRates
}
