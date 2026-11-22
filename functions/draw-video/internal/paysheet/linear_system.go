package paysheet

import (
	"errors"
	"math"
)

var errNoUniqueSolution = errors.New("linear system has no unique solution")

func solveLinearSystem(coefficients [][]float64, constants []float64) ([]float64, error) {
	size := len(constants)
	equations := make([][]float64, size)
	for i := range equations {
		equations[i] = make([]float64, size+1)
		copy(equations[i], coefficients[i])
		equations[i][size] = constants[i]
	}
	for column := range size {
		largest := column
		for row := column + 1; row < size; row++ {
			if math.Abs(equations[row][column]) > math.Abs(equations[largest][column]) {
				largest = row
			}
		}
		if equations[largest][column] == 0 {
			return nil, errNoUniqueSolution
		}
		equations[column], equations[largest] = equations[largest], equations[column]
		for row := column + 1; row < size; row++ {
			factor := equations[row][column] / equations[column][column]
			for i := column; i <= size; i++ {
				equations[row][i] -= factor * equations[column][i]
			}
		}
	}
	solution := make([]float64, size)
	for row := size - 1; row >= 0; row-- {
		remainder := equations[row][size]
		for column := row + 1; column < size; column++ {
			remainder -= equations[row][column] * solution[column]
		}
		solution[row] = remainder / equations[row][row]
	}
	return solution, nil
}
