package domain

// JackpotIncrement calcula el aporte al jackpot a partir de un conjunto de
// tickets: suma únicamente las utilidades positivas de la casa (apostado -
// pagado) y aplica una sola vez el porcentaje del pozo, para minimizar la
// pérdida por redondeo entero.
func JackpotIncrement(jackpot Jackpot, tickets []Ticket, balls Bitmask) int64 {
	var totalProfit int64
	for _, ticket := range tickets {
		if profit := ticket.HouseProfit(balls); profit > 0 {
			totalProfit += profit
		}
	}
	return jackpot.IncrementFor(totalProfit)
}
