package domain

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
)

func TestAdminBetsAddStakesAndPrizes(t *testing.T) {
	adminBets := AdminBets{Bets: []Bet{
		{Numbers: bitmask.MaskOf([]int{7}), StakeCents: 200},
		{Numbers: bitmask.MaskOf([]int{8, 50}), StakeCents: 100},
		{Numbers: bitmask.MaskOf([]int{60, 61, 62}), StakeCents: 400},
	}}
	drawn := bitmask.MaskOf([]int{7, 8, 9})
	if got := adminBets.StakeCents(); got != 700 {
		t.Fatalf("StakeCents = %d, want 700", got)
	}
	if got := adminBets.PrizeCents(drawn); got != 800 {
		t.Fatalf("PrizeCents = %d, want 800", got)
	}
}

func TestPayoutRatesDividePrizesBySales(t *testing.T) {
	adminBets := AdminBets{Bets: []Bet{
		{Numbers: bitmask.MaskOf([]int{1, 2}), StakeCents: 400},
		{Numbers: bitmask.MaskOf([]int{3, 4}), StakeCents: 400},
		{Numbers: bitmask.MaskOf([]int{5, 6, 7}), StakeCents: 100},
		{Numbers: bitmask.MaskOf([]int{8, 9}), StakeCents: 100},
	}}
	videos := []bitmask.Mask{
		bitmask.MaskOf([]int{1}),
		bitmask.MaskOf([]int{1, 3}),
		bitmask.MaskOf([]int{1, 3, 5, 6}),
		bitmask.MaskOf([]int{1, 5, 6, 8, 9}),
	}
	got := adminBets.PayoutRates(videos)
	want := []float64{0.4, 0.8, 1.0, 1.6}
	if !slices.Equal(got, want) {
		t.Fatalf("PayoutRates = %v, want %v", got, want)
	}
}

func TestPayoutRatesMatchSlowRates(t *testing.T) {
	random := rand.New(rand.NewPCG(11, 12))
	for range 200 {
		ballsByVideo := make([][]int, 50)
		videos := make([]bitmask.Mask, len(ballsByVideo))
		for i := range ballsByVideo {
			ballsByVideo[i] = randomNumbers(random, 20)
			videos[i] = bitmask.MaskOf(ballsByVideo[i])
		}
		var adminBets AdminBets
		var stakeCents int64
		prizeCentsByVideo := make([]int64, len(ballsByVideo))
		for range 1 + random.IntN(30) {
			numbers := randomNumbers(random, 1+random.IntN(10))
			betStakeCents := int64(1 + random.IntN(1000000))
			adminBets.Bets = append(adminBets.Bets, Bet{Numbers: bitmask.MaskOf(numbers), StakeCents: betStakeCents})
			stakeCents += betStakeCents
			for i, balls := range ballsByVideo {
				prizeCentsByVideo[i] += slowPrizeCents(numbers, betStakeCents, balls)
			}
		}
		want := make([]float64, len(ballsByVideo))
		for i, prizeCents := range prizeCentsByVideo {
			want[i] = float64(prizeCents) / float64(stakeCents)
		}
		got := adminBets.PayoutRates(videos)
		if !slices.Equal(got, want) {
			t.Fatalf("bets %+v: PayoutRates = %v, want %v", adminBets.Bets, got, want)
		}
	}
}

func TestAdminBetsMatchSlowTotals(t *testing.T) {
	random := rand.New(rand.NewPCG(9, 10))
	for range 2000 {
		balls := randomNumbers(random, 20)
		var adminBets AdminBets
		var wantStakeCents, wantPrizeCents int64
		for range 1 + random.IntN(30) {
			numbers := randomNumbers(random, 1+random.IntN(10))
			stakeCents := int64(1 + random.IntN(1000000))
			adminBets.Bets = append(adminBets.Bets, Bet{Numbers: bitmask.MaskOf(numbers), StakeCents: stakeCents})
			wantStakeCents += stakeCents
			wantPrizeCents += slowPrizeCents(numbers, stakeCents, balls)
		}
		if got := adminBets.StakeCents(); got != wantStakeCents {
			t.Fatalf("bets %+v: StakeCents = %d, want %d", adminBets.Bets, got, wantStakeCents)
		}
		if got := adminBets.PrizeCents(bitmask.MaskOf(balls)); got != wantPrizeCents {
			t.Fatalf("bets %+v balls %v: PrizeCents = %d, want %d", adminBets.Bets, balls, got, wantPrizeCents)
		}
	}
}
