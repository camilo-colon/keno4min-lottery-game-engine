package domain

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
)

const randomRounds = 10000

func slowPrizeCents(numbers []int, stakeCents int64, balls []int) int64 {
	hits := 0
	for _, number := range numbers {
		if slices.Contains(balls, number) {
			hits++
		}
	}
	return stakeCents * updateTicketsPaymentTable[len(numbers)][hits] / 100
}

func randomNumbers(random *rand.Rand, size int) []int {
	shuffled := random.Perm(80)
	numbers := make([]int, size)
	for i := range numbers {
		numbers[i] = shuffled[i] + 1
	}
	return numbers
}

func TestNewKeepsNumbersAndStake(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
	}{
		{name: "one number", numbers: []int{80}},
		{name: "ten numbers", numbers: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numbers := bitmask.MaskOf(tt.numbers)
			got, err := New(numbers, 500)
			if err != nil {
				t.Fatalf("New returned error: %v", err)
			}
			want := Bet{Numbers: numbers, StakeCents: 500}
			if got != want {
				t.Fatalf("New = %+v, want %+v", got, want)
			}
		})
	}
}

func TestNewRejectsInvalidBets(t *testing.T) {
	tests := []struct {
		name       string
		numbers    []int
		stakeCents int64
		wantErr    string
	}{
		{name: "no numbers", numbers: []int{}, stakeCents: 500, wantErr: "bet has 0 numbers"},
		{name: "eleven numbers", numbers: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, stakeCents: 500, wantErr: "bet has 11 numbers"},
		{name: "zero stake", numbers: []int{7}, stakeCents: 0, wantErr: "bet stake is 0 cents"},
		{name: "negative stake", numbers: []int{7}, stakeCents: -500, wantErr: "bet stake is -500 cents"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(bitmask.MaskOf(tt.numbers), tt.stakeCents)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("New error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestPrizeOfOneNumberHitIsThreeAndAHalfTimesTheStake(t *testing.T) {
	bet := Bet{Numbers: bitmask.MaskOf([]int{7}), StakeCents: 200}
	drawn := bitmask.MaskOf([]int{7, 8, 9})
	got := bet.PrizeCents(drawn)
	if got != 700 {
		t.Fatalf("PrizeCents = %d, want 700", got)
	}
}

func TestPrizeDropsFractionsOfACent(t *testing.T) {
	bet := Bet{Numbers: bitmask.MaskOf([]int{7}), StakeCents: 1}
	drawn := bitmask.MaskOf([]int{7, 8, 9})
	got := bet.PrizeCents(drawn)
	if got != 3 {
		t.Fatalf("PrizeCents = %d, want 3", got)
	}
}

func TestPrizeMatchesSlowPrize(t *testing.T) {
	random := rand.New(rand.NewPCG(5, 6))
	for range randomRounds {
		numbers := randomNumbers(random, 1+random.IntN(10))
		balls := randomNumbers(random, 20)
		stakeCents := int64(1 + random.IntN(1000000))
		bet := Bet{Numbers: bitmask.MaskOf(numbers), StakeCents: stakeCents}
		got := bet.PrizeCents(bitmask.MaskOf(balls))
		want := slowPrizeCents(numbers, stakeCents, balls)
		if got != want {
			t.Fatalf("numbers %v stake %d balls %v: PrizeCents = %d, want %d", numbers, stakeCents, balls, got, want)
		}
	}
}
