package utils

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// RandomIndex retorna un índice aleatorio entre [0, n).
// Usa crypto/rand para que sea criptográficamente seguro.
func RandomIndex(n int) int {
	if n <= 1 {
		return 0
	}
	var b [8]byte
	rand.Read(b[:])
	return int(binary.LittleEndian.Uint64(b[:]) % uint64(n))
}

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
