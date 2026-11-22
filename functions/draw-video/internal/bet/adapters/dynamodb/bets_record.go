package dynamodb

import (
	"encoding/binary"
	"fmt"

	bet "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
)

func decodeBets(record []byte) ([]bet.Bet, error) {
	if len(record) == 0 || len(record)%24 != 0 {
		return nil, fmt.Errorf("bets record has %d bytes, want a positive multiple of 24", len(record))
	}
	bets := make([]bet.Bet, 0, len(record)/24)
	for offset := 0; offset < len(record); offset += 24 {
		encoded := record[offset : offset+24]
		numbers := bitmask.Mask{
			Low:  binary.LittleEndian.Uint64(encoded[0:8]),
			High: binary.LittleEndian.Uint64(encoded[8:16]),
		}
		stakeCents := int64(binary.LittleEndian.Uint64(encoded[16:24]))
		decoded, err := bet.New(numbers, stakeCents)
		if err != nil {
			return nil, fmt.Errorf("bet at byte %d: %w", offset, err)
		}
		bets = append(bets, decoded)
	}
	return bets, nil
}
