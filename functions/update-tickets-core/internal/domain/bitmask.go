package domain

import "math/bits"

// Bitmask representa la máscara de bits de un conjunto de números (1..80).
type Bitmask struct {
	Mask1 int64 `bson:"mask1" json:"mask1"`
	Mask2 int64 `bson:"mask2" json:"mask2"`
}

// NewBitmask construye la máscara de un conjunto de números (1..80). El número n
// ocupa el bit n-1: 1..64 caen en Mask1 y 65..80 en Mask2. El core no persiste
// el bitmask de cada apuesta, así que hay que reconstruirlo desde los números.
// Números fuera del rango 1..80 se ignoran.
func NewBitmask(nums []int64) Bitmask {
	var b Bitmask
	for _, n := range nums {
		if n < 1 || n > 80 {
			continue
		}
		pos := n - 1
		if pos < 64 {
			b.Mask1 |= int64(1) << uint(pos)
		} else {
			b.Mask2 |= int64(1) << uint(pos-64)
		}
	}
	return b
}

// Count devuelve cuántos números contiene la máscara.
func (b Bitmask) Count() int {
	return bits.OnesCount64(uint64(b.Mask1)) + bits.OnesCount64(uint64(b.Mask2))
}

// Matches devuelve cuántos números de la máscara coinciden con otra.
func (b Bitmask) Matches(other Bitmask) int {
	return bits.OnesCount64(uint64(b.Mask1&other.Mask1)) +
		bits.OnesCount64(uint64(b.Mask2&other.Mask2))
}
