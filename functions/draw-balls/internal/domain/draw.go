package domain

import (
	"fmt"
	"strconv"
	"strings"
)

type Draws struct {
	Id   string `bson:"_id,omitempty" json:"id"`
	Game string `bson:"game" json:"game"`
	Idv  string `bson:"idv" json:"idv"`
}

func (d *Draws) ToGameBalls() (*GameBalls, error) {
	balls, err := parseBallsFromIDV(d.Idv)
	if err != nil {
		return nil, err
	}

	mask, err := BitMaskFromNumbers(balls)
	if err != nil {
		return nil, err
	}

	return &GameBalls{
		Nums: balls,
		Mask: mask,
	}, nil
}

func parseBallsFromIDV(idv string) ([]uint64, error) {
	return parseRoundPrefixedIDV(idv)
}

func parseRoundPrefixedIDV(idv string) ([]uint64, error) {
	parts := strings.SplitN(idv, "_", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("missing round separator")
	}

	round := strings.TrimSpace(parts[0])
	if round == "" {
		return nil, fmt.Errorf("missing round prefix")
	}
	if _, err := strconv.ParseUint(round, 10, 64); err != nil {
		return nil, fmt.Errorf("invalid round prefix %q: %w", round, err)
	}

	return parseRawBalls(parts[1])
}

func parseRawBalls(raw string) ([]uint64, error) {
	rawBalls := strings.Split(raw, ",")
	if len(rawBalls) != 20 {
		return nil, fmt.Errorf("expected 20 balls, got %d", len(rawBalls))
	}

	balls := make([]uint64, 0, len(rawBalls))
	for _, rawBall := range rawBalls {
		n, err := strconv.ParseUint(strings.TrimSpace(rawBall), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ball %q: %w", rawBall, err)
		}
		balls = append(balls, n)
	}

	return balls, nil
}
