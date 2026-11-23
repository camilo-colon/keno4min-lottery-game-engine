package dynamodb

import (
	"encoding/binary"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	bet "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
)

func encodeBets(bets []bet.Bet) []byte {
	var record []byte
	for _, decoded := range bets {
		record = binary.LittleEndian.AppendUint64(record, decoded.Numbers.Low)
		record = binary.LittleEndian.AppendUint64(record, decoded.Numbers.High)
		record = binary.LittleEndian.AppendUint64(record, uint64(decoded.StakeCents))
	}
	return record
}

func randomBet(random *rand.Rand) bet.Bet {
	numbers := random.Perm(80)[:1+random.IntN(10)]
	for i := range numbers {
		numbers[i]++
	}
	return bet.Bet{Numbers: bitmask.MaskOf(numbers), StakeCents: int64(1 + random.IntN(1000000))}
}

func TestDecodeBetsFollowsBetsRecordLayout(t *testing.T) {
	record := []byte{
		0x01, 0, 0, 0, 0, 0, 0, 0x80,
		0x01, 0x80, 0, 0, 0, 0, 0, 0,
		0xF4, 0x01, 0, 0, 0, 0, 0, 0,
	}
	got, err := decodeBets(record)
	if err != nil {
		t.Fatalf("decodeBets returned error: %v", err)
	}
	want := []bet.Bet{{Numbers: bitmask.MaskOf([]int{1, 64, 65, 80}), StakeCents: 500}}
	if !slices.Equal(got, want) {
		t.Fatalf("decodeBets = %+v, want %+v", got, want)
	}
}

func TestDecodeBetsReturnsEveryEncodedBet(t *testing.T) {
	random := rand.New(rand.NewPCG(7, 8))
	for range 1000 {
		want := make([]bet.Bet, 1+random.IntN(15))
		for i := range want {
			want[i] = randomBet(random)
		}
		got, err := decodeBets(encodeBets(want))
		if err != nil {
			t.Fatalf("bets %+v: decodeBets returned error: %v", want, err)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("decodeBets = %+v, want %+v", got, want)
		}
	}
}

func TestDecodeBetsRejectsMalformedRecords(t *testing.T) {
	valid := bet.Bet{Numbers: bitmask.MaskOf([]int{7}), StakeCents: 500}
	noNumbers := bet.Bet{StakeCents: 500}
	elevenNumbers := bet.Bet{Numbers: bitmask.MaskOf([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}), StakeCents: 500}
	zeroStake := bet.Bet{Numbers: bitmask.MaskOf([]int{7})}
	negativeStake := bet.Bet{Numbers: bitmask.MaskOf([]int{7}), StakeCents: -500}
	tests := []struct {
		name    string
		record  []byte
		wantErr string
	}{
		{name: "empty record", record: []byte{}, wantErr: "bets record has 0 bytes"},
		{name: "one byte short", record: encodeBets([]bet.Bet{valid})[:23], wantErr: "bets record has 23 bytes"},
		{name: "one byte extra", record: append(encodeBets([]bet.Bet{valid}), 0), wantErr: "bets record has 25 bytes"},
		{name: "bet without numbers", record: encodeBets([]bet.Bet{noNumbers}), wantErr: "bet at byte 0: bet has 0 numbers"},
		{name: "second bet with eleven numbers", record: encodeBets([]bet.Bet{valid, elevenNumbers}), wantErr: "bet at byte 24: bet has 11 numbers"},
		{name: "zero stake", record: encodeBets([]bet.Bet{zeroStake}), wantErr: "bet at byte 0: bet stake is 0 cents"},
		{name: "negative stake", record: encodeBets([]bet.Bet{negativeStake}), wantErr: "bet at byte 0: bet stake is -500 cents"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeBets(tt.record)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("decodeBets error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
