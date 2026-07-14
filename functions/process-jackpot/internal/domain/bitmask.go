package domain

import "math/bits"

// Bitmask representa la máscara de bits de un conjunto de números (1..80).
type Bitmask struct {
	Mask1 int64 `bson:"mask1" json:"mask1"`
	Mask2 int64 `bson:"mask2" json:"mask2"`
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
