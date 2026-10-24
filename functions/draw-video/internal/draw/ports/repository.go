package ports

import (
	"context"

	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
)

type Repository interface {
	Update(ctx context.Context, d *draw.Draw) error
}
