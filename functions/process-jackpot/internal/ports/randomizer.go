package ports

// Randomizer abstrae la fuente de aleatoriedad para poder testear el sorteo del
// jackpot de forma determinista.
type Randomizer interface {
	// Intn devuelve un entero aleatorio en [0, n).
	Intn(n int) int
	// Int64Between devuelve un entero aleatorio en [min, max].
	Int64Between(min, max int64) int64
}
