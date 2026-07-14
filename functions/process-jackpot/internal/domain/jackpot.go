package domain

// Jackpot es un value object: la instancia viva de un pozo acumulado
// (p. ej. Club.JP1). No tiene identidad propia; vive embebido en el club.
type Jackpot struct {
	Percent int64 `bson:"percent" json:"percent"`
	Target  int64 `bson:"target" json:"target"`
	Value   int64 `bson:"value" json:"value"`
	Min     int64 `bson:"min" json:"min"`
	Max     int64 `bson:"max" json:"max"`
}

// IncrementFor convierte una utilidad en el aporte al pozo: el porcentaje del
// jackpot aplicado sobre esa utilidad. Una utilidad no positiva no aporta nada.
func (j Jackpot) IncrementFor(profit int64) int64 {
	if profit <= 0 {
		return 0
	}
	return profit * j.Percent / 100
}

// ShouldPlay indica si el pozo debe jugarse: cuando alcanzó (o superó) su target.
func (j Jackpot) ShouldPlay() bool {
	return j.Value >= j.Target
}

// NewJackpotFromConfig construye un jackpot fresco a partir de la config del club
// tras jugarse el pozo: el value arranca en la semilla y el target es un valor
// aleatorio dentro de [min, max] provisto por el llamador.
func NewJackpotFromConfig(cfg JackpotConfig, target int64) Jackpot {
	return Jackpot{
		Percent: cfg.Percent,
		Target:  target,
		Value:   cfg.Seed,
		Min:     cfg.Min,
		Max:     cfg.Max,
	}
}
