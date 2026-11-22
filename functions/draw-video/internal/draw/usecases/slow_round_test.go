package usecases

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	bet "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
	video "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/domain"
)

var updateTicketsPaymentTable = map[int][]int64{
	1:  {0, 350},
	2:  {0, 100, 1000},
	3:  {0, 0, 200, 5000},
	4:  {0, 0, 100, 1000, 10000},
	5:  {0, 0, 100, 300, 2000, 15000},
	6:  {0, 0, 0, 200, 1500, 6000, 50000},
	7:  {100, 0, 0, 200, 400, 2000, 8000, 100000},
	8:  {100, 0, 0, 0, 500, 1500, 5000, 20000, 200000},
	9:  {200, 0, 0, 0, 100, 1000, 2500, 12500, 100000, 500000},
	10: {200, 0, 0, 0, 0, 500, 3000, 10000, 30000, 200000, 1000000},
}

type slowBet struct {
	numbers    []int
	stakeCents int64
}

func slowPrizeCents(slow slowBet, balls []int) int64 {
	hits := 0
	for _, number := range slow.numbers {
		for _, ball := range balls {
			if number == ball {
				hits++
			}
		}
	}
	return slow.stakeCents * updateTicketsPaymentTable[len(slow.numbers)][hits] / 100
}

func slowPayoutRates(slowBets []slowBet, ballsByVideo [][]int) []float64 {
	var stakeCents int64
	for _, slow := range slowBets {
		stakeCents += slow.stakeCents
	}
	payoutRates := make([]float64, len(ballsByVideo))
	for video, balls := range ballsByVideo {
		var prizeCents int64
		for _, slow := range slowBets {
			prizeCents += slowPrizeCents(slow, balls)
		}
		payoutRates[video] = float64(prizeCents) / float64(stakeCents)
	}
	return payoutRates
}

func determinant(matrix [][]float64) float64 {
	if len(matrix) == 1 {
		return matrix[0][0]
	}
	var total float64
	sign := 1.0
	for column := range matrix {
		smaller := make([][]float64, 0, len(matrix)-1)
		for _, row := range matrix[1:] {
			smaller = append(smaller, append(slices.Clone(row[:column]), row[column+1:]...))
		}
		total += sign * matrix[0][column] * determinant(smaller)
		sign = -sign
	}
	return total
}

func slowSolve(coefficients [][]float64, constants []float64) []float64 {
	solution := make([]float64, len(constants))
	for unknown := range constants {
		replaced := make([][]float64, len(coefficients))
		for row := range coefficients {
			replaced[row] = slices.Clone(coefficients[row])
			replaced[row][unknown] = constants[row]
		}
		solution[unknown] = determinant(replaced) / determinant(coefficients)
	}
	return solution
}

func slowWeights(payoutRatesByAdmin [][]float64) []float64 {
	admins := len(payoutRatesByAdmin)
	videos := len(payoutRatesByAdmin[0])
	averages := make([]float64, admins)
	for admin := range admins {
		for video := range videos {
			averages[admin] += payoutRatesByAdmin[admin][video] / float64(videos)
		}
	}
	effects := make([][]float64, admins)
	excessOverTarget := make([]float64, admins)
	for admin := range admins {
		effects[admin] = make([]float64, admins)
		for other := range admins {
			for video := range videos {
				difference := payoutRatesByAdmin[admin][video] - averages[admin]
				otherDifference := payoutRatesByAdmin[other][video] - averages[other]
				effects[admin][other] += difference * otherDifference / float64(videos)
			}
		}
		excessOverTarget[admin] = averages[admin] - 0.90
	}
	tilts := slowSolve(effects, excessOverTarget)
	weights := make([]float64, videos)
	for video := range videos {
		weights[video] = 1
		for admin := range admins {
			weights[video] -= tilts[admin] * (payoutRatesByAdmin[admin][video] - averages[admin])
		}
		if weights[video] < 0 {
			weights[video] = 0
		}
	}
	return weights
}

func randomNumbers(random *rand.Rand, size int) []int {
	shuffled := random.Perm(80)
	numbers := make([]int, size)
	for i := range numbers {
		numbers[i] = shuffled[i] + 1
	}
	return numbers
}

func TestVideoWeightsMatchSlowRound(t *testing.T) {
	random := rand.New(rand.NewPCG(19, 20))
	roundsWithCuts := 0
	for range 2000 {
		ballsByVideo := make([][]int, 100)
		videos := make([]video.Video, len(ballsByVideo))
		for i := range ballsByVideo {
			ballsByVideo[i] = randomNumbers(random, 20)
			videos[i] = video.Video{Balls: ballsByVideo[i]}
		}
		betsByAdmin := make([]bet.AdminBets, 1+random.IntN(4))
		slowPayoutRatesByAdmin := make([][]float64, len(betsByAdmin))
		for admin := range betsByAdmin {
			var slowBets []slowBet
			for range 1 + random.IntN(8) {
				slow := slowBet{numbers: randomNumbers(random, 1+random.IntN(10)), stakeCents: int64(100 * (1 + random.IntN(50)))}
				slowBets = append(slowBets, slow)
				betsByAdmin[admin].Bets = append(betsByAdmin[admin].Bets, bet.Bet{Numbers: bitmask.MaskOf(slow.numbers), StakeCents: slow.stakeCents})
			}
			slowPayoutRatesByAdmin[admin] = slowPayoutRates(slowBets, ballsByVideo)
		}
		got, err := videoWeights(betsByAdmin, videos)
		if err != nil {
			t.Fatalf("bets %+v: videoWeights returned error: %v", betsByAdmin, err)
		}
		want := slowWeights(slowPayoutRatesByAdmin)
		for i := range want {
			if math.Abs(got[i]-want[i]) > 1e-6 {
				t.Fatalf("bets %+v: weight of video %d = %v, want %v", betsByAdmin, i, got[i], want[i])
			}
		}
		if slices.Contains(got, 0) {
			roundsWithCuts++
		}
	}
	if roundsWithCuts < 100 {
		t.Fatalf("only %d rounds had cut weights, want at least 100", roundsWithCuts)
	}
}
