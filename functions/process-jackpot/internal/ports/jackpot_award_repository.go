package ports

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

// JackpotAwardRepository persiste el histórico de jackpots ganados.
type JackpotAwardRepository interface {
	// Record guarda un premio otorgado. Debe ejecutarse dentro de la misma
	// transacción que el pago y el reset del pozo, para que el histórico sea
	// consistente con el movimiento real de dinero.
	Record(ctx context.Context, award domain.JackpotAward) error
}
