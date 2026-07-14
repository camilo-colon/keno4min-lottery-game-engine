package domain

// JackpotIncrement calcula el aporte al jackpot a partir de la utilidad NETA de la
// casa sobre todos los tickets del club en el juego: la sumatoria de
// (apostado - pagado) de cada ticket, donde las pérdidas RESTAN. Solo si el neto
// es positivo se aplica el porcentaje del pozo (una sola división, al final).
func JackpotIncrement(jackpot Jackpot, tickets []Ticket, balls Bitmask) int64 {
	var netProfit int64
	for _, ticket := range tickets {
		netProfit += ticket.HouseProfit(balls)
	}
	return jackpot.IncrementFor(netProfit)
}
