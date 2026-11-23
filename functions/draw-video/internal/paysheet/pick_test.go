package paysheet

import (
	"errors"
	"math"
	"testing"
)

func TestVideoAtFollowsTheAccumulatedWeights(t *testing.T) {
	weights := []float64{1.1466666666666667, 1.04, 0.9866666666666667, 0.8266666666666667}
	tests := []struct {
		name     string
		fraction float64
		want     int
	}{
		{name: "start of the first video", fraction: 0, want: 0},
		{name: "inside the first video", fraction: 0.28, want: 0},
		{name: "just past the first video", fraction: 0.29, want: 1},
		{name: "two and a half of four", fraction: 0.625, want: 2},
		{name: "end of the last video", fraction: 0.999999, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := videoAt(weights, tt.fraction)
			if err != nil || got != tt.want {
				t.Fatalf("videoAt(%v) = %d, %v; want %d", tt.fraction, got, err, tt.want)
			}
		})
	}
}

func TestVideoAtNeverPicksAVideoWithoutWeight(t *testing.T) {
	weights := []float64{0, 2, 0, 1, 0}
	for step := range 1000 {
		fraction := float64(step) / 1000
		got, err := videoAt(weights, fraction)
		if err != nil || weights[got] == 0 {
			t.Fatalf("videoAt(%v) = %d, %v; want a video with weight", fraction, got, err)
		}
	}
}

func TestVideoAtRejectsUnusableWeights(t *testing.T) {
	tests := []struct {
		name    string
		weights []float64
	}{
		{name: "no videos", weights: nil},
		{name: "every weight is zero", weights: []float64{0, 0, 0}},
		{name: "negative total", weights: []float64{1, -2}},
		{name: "not a number", weights: []float64{1, math.NaN()}},
		{name: "infinite", weights: []float64{1, math.Inf(1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := videoAt(tt.weights, 0.5)
			if !errors.Is(err, errUnusableWeights) {
				t.Fatalf("videoAt error = %v, want errUnusableWeights", err)
			}
		})
	}
}

func TestPickVideoFollowsTheProbabilities(t *testing.T) {
	weights := []float64{1.1466666666666667, 1.04, 0.9866666666666667, 0.8266666666666667}
	wantProbabilities := []float64{0.2866666666666667, 0.26, 0.2466666666666667, 0.2066666666666667}
	picks := 100000
	timesPicked := make([]int, len(weights))
	for range picks {
		video, err := PickVideo(weights)
		if err != nil {
			t.Fatalf("PickVideo returned error: %v", err)
		}
		timesPicked[video]++
	}
	for video, want := range wantProbabilities {
		got := float64(timesPicked[video]) / float64(picks)
		if math.Abs(got-want) > 0.01 {
			t.Fatalf("video %d was picked %v of the time, want about %v", video, got, want)
		}
	}
}
