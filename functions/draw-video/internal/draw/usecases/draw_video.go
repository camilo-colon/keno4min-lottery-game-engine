package usecases

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
	drawports "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/ports"
	videoports "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/ports"
)

var ErrNoVideos = errors.New("no videos available")

type DrawVideo struct {
	draws  drawports.Repository
	videos videoports.Repository
}

func NewDrawVideo(draws drawports.Repository, videos videoports.Repository) *DrawVideo {
	return &DrawVideo{draws: draws, videos: videos}
}

func (u *DrawVideo) Execute(ctx context.Context, input *draw.Draw) (*draw.Draw, error) {
	videos, err := u.videos.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get videos: %w", err)
	}
	if len(videos) == 0 {
		return nil, ErrNoVideos
	}
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(videos))))
	if err != nil {
		return nil, fmt.Errorf("select video: %w", err)
	}
	selected := videos[index.Int64()]
	result := *input
	result.AssignVideo(selected.ID, selected.Balls)
	if err := u.draws.Update(ctx, &result); err != nil {
		return nil, fmt.Errorf("update draw video: %w", err)
	}
	return &result, nil
}
