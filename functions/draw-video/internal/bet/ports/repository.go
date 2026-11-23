package ports

import (
	"context"

	bet "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/domain"
)

type Repository interface {
	GetByTIR(ctx context.Context, tir uint64) ([]bet.AdminBets, error)
}
