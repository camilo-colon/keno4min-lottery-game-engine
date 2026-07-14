package adapters

import (
	"crypto/rand"
	"math/big"
)

// CryptoRandomizer implementa ports.Randomizer con aleatoriedad criptográfica,
// apropiada para sortear premios (dinero).
type CryptoRandomizer struct{}

func NewCryptoRandomizer() CryptoRandomizer {
	return CryptoRandomizer{}
}

// Intn devuelve un entero aleatorio en [0, n).
func (CryptoRandomizer) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic("crypto/rand failure: " + err.Error())
	}
	return int(v.Int64())
}

// Int64Between devuelve un entero aleatorio en [min, max].
func (CryptoRandomizer) Int64Between(min, max int64) int64 {
	if max <= min {
		return min
	}
	span := max - min + 1 // inclusivo en ambos extremos
	v, err := rand.Int(rand.Reader, big.NewInt(span))
	if err != nil {
		panic("crypto/rand failure: " + err.Error())
	}
	return min + v.Int64()
}
