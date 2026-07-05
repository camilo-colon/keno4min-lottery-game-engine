package service

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
)

type Draw struct {
	Balls   []uint64
	BitMask domain.BitMask
}

func GenerateDraws(totalDraws int) ([]Draw, error) {
	allDraws := make([]Draw, 0, totalDraws)
	for range totalDraws {
		draw, err := GenerateDraw()
		if err != nil {
			return nil, fmt.Errorf("error generating draws: %w", err)
		}
		allDraws = append(allDraws, Draw{
			Balls:   draw.Balls,
			BitMask: draw.BitMask,
		})
	}
	return allDraws, nil
}

// GenerateBalls genera 20 balotas aleatorias del 1 al 80 usando crypto/rand
func GenerateBalls() []uint64 {
	selected := make(map[uint64]bool)
	const RANGE = 80
	const MAX_UINT32 = uint64(1) << 32 // 2^32 = 4294967296
	const MAX_VALID = uint32(MAX_UINT32 - (MAX_UINT32 % RANGE))
	for len(selected) < 20 {
		buffer := make([]byte, 4)
		_, err := rand.Read(buffer)
		if err != nil {
			panic(fmt.Sprintf("Error generating random bytes: %v", err))
		}
		random := binary.BigEndian.Uint32(buffer)

		if random >= MAX_VALID {
			continue
		}

		number := uint64(random%RANGE) + 1
		selected[number] = true
	}
	result := make([]uint64, 0, len(selected))
	for number := range selected {
		result = append(result, number)
	}
	return result
}

// GenerateDraw genera las balotas y construye el GameBalls con nums y mask
func GenerateDraw() (Draw, error) {
	nums := GenerateBalls()

	mask, err := domain.BitMaskFromNumbers(nums)
	if err != nil {
		return Draw{}, fmt.Errorf("error creating bitmask from numbers: %w", err)
	}

	return Draw{
		Balls:   nums,
		BitMask: mask,
	}, nil
}
