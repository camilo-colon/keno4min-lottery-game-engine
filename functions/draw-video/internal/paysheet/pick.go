package paysheet

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
)

var errUnusableWeights = errors.New("video weights do not add up to a positive number")

func PickVideo(weights []float64) (int, error) {
	fraction, err := randomFraction()
	if err != nil {
		return 0, err
	}
	return videoAt(weights, fraction)
}

func randomFraction() (float64, error) {
	steps := big.NewInt(1 << 53)
	step, err := rand.Int(rand.Reader, steps)
	if err != nil {
		return 0, fmt.Errorf("read random number: %w", err)
	}
	return float64(step.Int64()) / float64(steps.Int64()), nil
}

func videoAt(weights []float64, fraction float64) (int, error) {
	var total float64
	for _, weight := range weights {
		total += weight
	}
	if math.IsNaN(total) || math.IsInf(total, 1) || total <= 0 {
		return 0, errUnusableWeights
	}
	var accumulated float64
	for video, weight := range weights {
		accumulated += weight
		if fraction*total < accumulated {
			return video, nil
		}
	}
	return 0, errUnusableWeights
}
