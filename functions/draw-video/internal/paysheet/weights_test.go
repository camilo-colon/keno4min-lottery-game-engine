package paysheet

import (
	"errors"
	"math/rand/v2"
	"slices"
	"testing"
)

func expectedPayoutRate(weights, payoutRates []float64) float64 {
	var weightedPayout, totalWeight float64
	for video, weight := range weights {
		weightedPayout += weight * payoutRates[video]
		totalWeight += weight
	}
	return weightedPayout / totalWeight
}

func TestWeightsOfOneAdmin(t *testing.T) {
	payoutRates := []float64{0.4, 0.8, 1.0, 1.6}
	weights, err := Weights([][]float64{payoutRates})
	if err != nil {
		t.Fatalf("Weights returned error: %v", err)
	}
	wantWeights := []float64{1.1466666666666667, 1.04, 0.9866666666666667, 0.8266666666666667}
	wantProbabilities := []float64{0.2866666666666667, 0.26, 0.2466666666666667, 0.2066666666666667}
	for video := range wantWeights {
		if !closeTo(weights[video], wantWeights[video]) {
			t.Fatalf("Weights = %v, want %v", weights, wantWeights)
		}
		if !closeTo(weights[video]/4, wantProbabilities[video]) {
			t.Fatalf("probability of video %d = %v, want %v", video, weights[video]/4, wantProbabilities[video])
		}
	}
	if got := 1000 * expectedPayoutRate(weights, payoutRates); !closeTo(got, 900) {
		t.Fatalf("expected payout = %v, want 900", got)
	}
}

func TestWeightsOfTwoAdmins(t *testing.T) {
	payoutRatesByAdmin := [][]float64{
		{0.4, 0.8, 1.0, 1.6},
		{1.4, 0.6, 1.2, 0.8},
	}
	weights, err := Weights(payoutRatesByAdmin)
	if err != nil {
		t.Fatalf("Weights returned error: %v", err)
	}
	for admin, payoutRates := range payoutRatesByAdmin {
		if got := expectedPayoutRate(weights, payoutRates); !closeTo(got, targetPayoutRate) {
			t.Fatalf("admin %d expects to pay %v, want %v", admin, got, targetPayoutRate)
		}
	}
}

func TestWeightsGiveEveryAdminTheTargetWhenNoWeightWasCut(t *testing.T) {
	random := rand.New(rand.NewPCG(15, 16))
	roundsWithoutCuts := 0
	for range 3000 {
		payoutRatesByAdmin := make([][]float64, 1+random.IntN(8))
		for admin := range payoutRatesByAdmin {
			payoutRatesByAdmin[admin] = make([]float64, 200)
			center := 0.85 + 0.1*random.Float64()
			for video := range payoutRatesByAdmin[admin] {
				payoutRatesByAdmin[admin][video] = center + random.Float64() - 0.5
			}
		}
		weights, err := Weights(payoutRatesByAdmin)
		if err != nil {
			t.Fatalf("Weights returned error: %v", err)
		}
		if slices.Contains(weights, 0) {
			continue
		}
		roundsWithoutCuts++
		for admin, payoutRates := range payoutRatesByAdmin {
			if got := expectedPayoutRate(weights, payoutRates); !closeTo(got, targetPayoutRate) {
				t.Fatalf("admin %d of %d expects to pay %v, want %v", admin, len(payoutRatesByAdmin), got, targetPayoutRate)
			}
		}
	}
	if roundsWithoutCuts < 2000 {
		t.Fatalf("only %d rounds had no cut weights, want at least 2000", roundsWithoutCuts)
	}
}

func TestWeightsAreNeverNegative(t *testing.T) {
	random := rand.New(rand.NewPCG(17, 18))
	roundsWithCuts := 0
	for range 1000 {
		payoutRatesByAdmin := make([][]float64, 1+random.IntN(8))
		for admin := range payoutRatesByAdmin {
			payoutRatesByAdmin[admin] = make([]float64, 200)
			center := 1.2 + 0.8*random.Float64()
			for video := range payoutRatesByAdmin[admin] {
				payoutRatesByAdmin[admin][video] = center + random.Float64() - 0.5
			}
		}
		weights, err := Weights(payoutRatesByAdmin)
		if err != nil {
			t.Fatalf("Weights returned error: %v", err)
		}
		for video, weight := range weights {
			if !(weight >= 0) {
				t.Fatalf("weight of video %d = %v, want it not negative", video, weight)
			}
		}
		if slices.Contains(weights, 0) {
			roundsWithCuts++
		}
	}
	if roundsWithCuts < 500 {
		t.Fatalf("only %d rounds had cut weights, want at least 500", roundsWithCuts)
	}
}

func TestWeightsRejectAdminsWithTheSamePayoutRates(t *testing.T) {
	payoutRates := []float64{0.4, 0.8, 1.0, 1.6}
	weights, err := Weights([][]float64{payoutRates, payoutRates})
	if weights != nil || !errors.Is(err, errNoUniqueSolution) {
		t.Fatalf("Weights = %v, %v; want errNoUniqueSolution", weights, err)
	}
}

func TestWeightsRejectARoundWithoutAdmins(t *testing.T) {
	weights, err := Weights(nil)
	if weights != nil || !errors.Is(err, errNoPayoutRates) {
		t.Fatalf("Weights = %v, %v; want errNoPayoutRates", weights, err)
	}
}
