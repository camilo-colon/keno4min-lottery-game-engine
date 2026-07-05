package domain

import (
	"fmt"
	"math/bits"
)

type BitMask struct {
	Mask1 int64 `bson:"mask1" json:"mask1"`
	Mask2 int64 `bson:"mask2" json:"mask2"`
}

func NewBitMask() BitMask {
	return BitMask{}
}

func FromMasks(mask1, mask2 int64) *BitMask {
	return &BitMask{
		Mask1: mask1,
		Mask2: mask2,
	}
}

func BitMaskFromNumbers(numbers []uint64) (BitMask, error) {
	bm := NewBitMask()
	for _, num := range numbers {
		if err := bm.Add(num); err != nil {
			return BitMask{}, err
		}
	}
	return bm, nil
}

func (b *BitMask) Add(num uint64) error {
	if num < 1 || num > 80 {
		return fmt.Errorf("number %d must be between 1 and 80", num)
	}

	bitPos := num - 1
	if bitPos < 64 {
		b.Mask1 |= int64(1) << bitPos
	} else {
		b.Mask2 |= int64(1) << (bitPos - 64)
	}

	return nil
}

func (b *BitMask) Count() int {
	return bits.OnesCount64(uint64(b.Mask1)) + bits.OnesCount64(uint64(b.Mask2))
}

func (b *BitMask) Intersection(other *BitMask) *BitMask {
	return &BitMask{
		Mask1: b.Mask1 & other.Mask1,
		Mask2: b.Mask2 & other.Mask2,
	}
}

func (b *BitMask) Matches(other *BitMask) int {
	return b.Intersection(other).Count()
}

func (b *BitMask) ToSlice() []uint64 {
	numbers := make([]uint64, 0, b.Count())

	tempMask := uint64(b.Mask1)
	for i := uint64(0); i < 64 && tempMask > 0; i++ {
		if tempMask&1 == 1 {
			numbers = append(numbers, i+1)
		}
		tempMask >>= 1
	}

	tempMask = uint64(b.Mask2)
	for i := uint64(0); i < 64 && tempMask != 0; i++ {
		if tempMask&1 == 1 {
			numbers = append(numbers, i+65)
		}
		tempMask >>= 1
	}

	return numbers
}
