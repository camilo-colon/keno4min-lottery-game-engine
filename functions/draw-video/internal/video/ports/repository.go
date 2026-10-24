package ports

import (
	"context"

	video "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/domain"
)

type Repository interface {
	GetAll(context.Context) ([]video.Video, error)
}
