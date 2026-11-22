package paysheet

import "errors"

const targetPayoutRate = 0.90

var errNoPayoutRates = errors.New("no payout rates to weigh the videos with")

func Weights(payoutRatesByAdmin [][]float64) ([]float64, error) {
	if len(payoutRatesByAdmin) == 0 {
		return nil, errNoPayoutRates
	}
	averages := make([]float64, len(payoutRatesByAdmin))
	differencesByAdmin := make([][]float64, len(payoutRatesByAdmin))
	for admin, payoutRates := range payoutRatesByAdmin {
		averages[admin] = average(payoutRates)
		differencesByAdmin[admin] = differencesFrom(payoutRates, averages[admin])
	}
	effects := make([][]float64, len(differencesByAdmin))
	excessOverTarget := make([]float64, len(differencesByAdmin))
	for admin, differences := range differencesByAdmin {
		effects[admin] = make([]float64, len(differencesByAdmin))
		for other, otherDifferences := range differencesByAdmin {
			effects[admin][other] = averageProduct(differences, otherDifferences)
		}
		excessOverTarget[admin] = averages[admin] - targetPayoutRate
	}
	tilts, err := solveLinearSystem(effects, excessOverTarget)
	if err != nil {
		return nil, err
	}
	weights := make([]float64, len(payoutRatesByAdmin[0]))
	for video := range weights {
		weight := 1.0
		for admin, differences := range differencesByAdmin {
			weight -= tilts[admin] * differences[video]
		}
		weights[video] = max(weight, 0)
	}
	return weights, nil
}

func average(values []float64) float64 {
	var total float64
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func differencesFrom(values []float64, reference float64) []float64 {
	differences := make([]float64, len(values))
	for i, value := range values {
		differences[i] = value - reference
	}
	return differences
}

func averageProduct(left, right []float64) float64 {
	var total float64
	for i := range left {
		total += left[i] * right[i]
	}
	return total / float64(len(left))
}
