package paysheet

import (
	"errors"
	"math"
	"math/rand/v2"
	"testing"
)

func closeTo(got, want float64) bool {
	return math.Abs(got-want) <= 1e-9
}

func TestSolveLinearSystemOfOneEquation(t *testing.T) {
	got, err := solveLinearSystem([][]float64{{0.1875}}, []float64{0.05})
	if err != nil {
		t.Fatalf("solveLinearSystem returned error: %v", err)
	}
	if len(got) != 1 || !closeTo(got[0], 0.26666666666666666) {
		t.Fatalf("solveLinearSystem = %v, want [0.2667]", got)
	}
}

func TestSolveLinearSystemOfTwoEquations(t *testing.T) {
	coefficients := [][]float64{
		{2, 1},
		{1, 3},
	}
	got, err := solveLinearSystem(coefficients, []float64{5, 10})
	if err != nil {
		t.Fatalf("solveLinearSystem returned error: %v", err)
	}
	if len(got) != 2 || !closeTo(got[0], 1) || !closeTo(got[1], 3) {
		t.Fatalf("solveLinearSystem = %v, want [1 3]", got)
	}
}

func TestSolveLinearSystemSatisfiesEveryEquation(t *testing.T) {
	random := rand.New(rand.NewPCG(13, 14))
	for range 5000 {
		size := 1 + random.IntN(12)
		coefficients := make([][]float64, size)
		constants := make([]float64, size)
		for row := range coefficients {
			coefficients[row] = make([]float64, size)
			for column := range coefficients[row] {
				coefficients[row][column] = 2*random.Float64() - 1
			}
			constants[row] = 2*random.Float64() - 1
		}
		solution, err := solveLinearSystem(coefficients, constants)
		if err != nil {
			t.Fatalf("coefficients %v constants %v: solveLinearSystem returned error: %v", coefficients, constants, err)
		}
		for row := range coefficients {
			var total float64
			for column, coefficient := range coefficients[row] {
				total += coefficient * solution[column]
			}
			if !closeTo(total, constants[row]) {
				t.Fatalf("coefficients %v constants %v: equation %d adds up to %v, want %v", coefficients, constants, row, total, constants[row])
			}
		}
	}
}

func TestSolveLinearSystemRejectsEquationsThatSayTheSame(t *testing.T) {
	coefficients := [][]float64{
		{1, 1},
		{1, 1},
	}
	got, err := solveLinearSystem(coefficients, []float64{10, 10})
	if got != nil || !errors.Is(err, errNoUniqueSolution) {
		t.Fatalf("solveLinearSystem = %v, %v; want errNoUniqueSolution", got, err)
	}
}
