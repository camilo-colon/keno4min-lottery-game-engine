package domain

// JackpotIncrement calcula el aporte al jackpot a partir de la utilidad NETA de la
// casa sobre todos los tickets del club en el juego: la sumatoria de
// (apostado - pagado) de cada ticket, donde las pérdidas RESTAN. Solo si el neto
// es positivo se aplica el porcentaje del pozo (una sola división, al final).
//
// El pagado (Win) ya viene persistido por la Lambda update-tickets, que corre
// entre DrawBalls y ProcessJackpot: este cálculo ya NO recalcula el payout
// contra las balotas.
func JackpotIncrement(jackpot Jackpot, tickets []Ticket) int64 {
	var netProfit int64
	for _, ticket := range tickets {
		netProfit += ticket.HouseProfit()
	}
	return jackpot.IncrementFor(netProfit)
}
