package bitmask

import "math/bits"

type Mask struct {
	Low  uint64
	High uint64
}

func MaskOf(numbers []int) Mask {
	var mask Mask
	for _, number := range numbers {
		if number <= 64 {
			mask.Low |= 1 << (number - 1)
		} else {
			mask.High |= 1 << (number - 65)
		}
	}
	return mask
}

func (m Mask) Count() int {
	return bits.OnesCount64(m.Low) + bits.OnesCount64(m.High)
}

func Hits(a, b Mask) int {
	return bits.OnesCount64(a.Low&b.Low) + bits.OnesCount64(a.High&b.High)
}
