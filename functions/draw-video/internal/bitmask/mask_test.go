package bitmask

import (
	"math/rand/v2"
	"testing"
)

const randomRounds = 10000

func slowHits(picked, drawn []int) int {
	hits := 0
	for _, pickedNumber := range picked {
		for _, drawnNumber := range drawn {
			if pickedNumber == drawnNumber {
				hits++
			}
		}
	}
	return hits
}

func randomNumbers(random *rand.Rand, size int) []int {
	shuffled := random.Perm(80)
	numbers := make([]int, size)
	for i := range numbers {
		numbers[i] = shuffled[i] + 1
	}
	return numbers
}

func TestMaskOfFollowsBetsRecordLayout(t *testing.T) {
	tests := []struct {
		name   string
		number int
		want   Mask
	}{
		{name: "number 1 is the first bit of Low", number: 1, want: Mask{Low: 1}},
		{name: "number 64 is the last bit of Low", number: 64, want: Mask{Low: 1 << 63}},
		{name: "number 65 is the first bit of High", number: 65, want: Mask{High: 1}},
		{name: "number 80 is bit 15 of High", number: 80, want: Mask{High: 1 << 15}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskOf([]int{tt.number})
			if got != tt.want {
				t.Fatalf("MaskOf(%d) = %+v, want %+v", tt.number, got, tt.want)
			}
		})
	}
}

func TestCountIsHowManyNumbersTheMaskHas(t *testing.T) {
	random := rand.New(rand.NewPCG(1, 2))
	for range randomRounds {
		numbers := randomNumbers(random, random.IntN(81))
		got := MaskOf(numbers).Count()
		if got != len(numbers) {
			t.Fatalf("numbers %v: Count() = %d, want %d", numbers, got, len(numbers))
		}
	}
}

func TestHitsMatchesSlowCount(t *testing.T) {
	random := rand.New(rand.NewPCG(3, 4))
	for range randomRounds {
		picked := randomNumbers(random, 1+random.IntN(10))
		drawn := randomNumbers(random, 20)
		got := Hits(MaskOf(picked), MaskOf(drawn))
		want := slowHits(picked, drawn)
		if got != want {
			t.Fatalf("picked %v drawn %v: Hits = %d, want %d", picked, drawn, got, want)
		}
	}
}
