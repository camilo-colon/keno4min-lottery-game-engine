package usecases

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	bet "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/domain"
	betports "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/ports"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
	drawports "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/ports"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/paysheet"
	video "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/domain"
	videoports "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/ports"
)

var ErrNoVideos = errors.New("no videos available")

type DrawVideo struct {
	draws  drawports.Repository
	videos videoports.Repository
	bets   betports.Repository
}

func NewDrawVideo(draws drawports.Repository, videos videoports.Repository, bets betports.Repository) *DrawVideo {
	return &DrawVideo{draws: draws, videos: videos, bets: bets}
}

func (u *DrawVideo) Execute(ctx context.Context, input *draw.Draw) (*draw.Draw, error) {
	videos, err := u.videos.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get videos: %w", err)
	}
	if len(videos) == 0 {
		return nil, ErrNoVideos
	}
	index, err := u.pickWithPaysheet(ctx, input.TIR, videos)
	if err != nil {
		slog.Warn("paysheet pick failed, picking uniformly", "tir", input.TIR, "error", err)
		index, err = pickUniformly(len(videos))
		if err != nil {
			return nil, fmt.Errorf("select video: %w", err)
		}
	}
	selected := videos[index]
	result := *input
	result.AssignVideo(selected.ID, selected.Balls)
	if err := u.draws.Update(ctx, &result); err != nil {
		return nil, fmt.Errorf("update draw video: %w", err)
	}
	return &result, nil
}

func (u *DrawVideo) pickWithPaysheet(ctx context.Context, tir uint64, videos []video.Video) (index int, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("paysheet panicked: %v", recovered)
		}
	}()
	readStart := time.Now()
	betsByAdmin, err := u.bets.GetByTIR(ctx, tir)
	if err != nil {
		return 0, fmt.Errorf("get bets: %w", err)
	}
	slog.Info("bets read", "tir", tir, "admins", len(betsByAdmin), "readTime", time.Since(readStart))
	if len(betsByAdmin) == 0 {
		return pickUniformly(len(videos))
	}
	computeStart := time.Now()
	weights, err := videoWeights(betsByAdmin, videos)
	if err != nil {
		return 0, fmt.Errorf("weigh videos: %w", err)
	}
	index, err = paysheet.PickVideo(weights)
	if err != nil {
		return 0, fmt.Errorf("pick video: %w", err)
	}
	slog.Info("paysheet computed", "tir", tir, "computeTime", time.Since(computeStart))
	return index, nil
}

func videoWeights(betsByAdmin []bet.AdminBets, videos []video.Video) ([]float64, error) {
	drawnByVideo := make([]bitmask.Mask, len(videos))
	for i, v := range videos {
		drawnByVideo[i] = bitmask.MaskOf(v.Balls)
	}
	payoutRatesByAdmin := make([][]float64, len(betsByAdmin))
	for i, adminBets := range betsByAdmin {
		payoutRatesByAdmin[i] = adminBets.PayoutRates(drawnByVideo)
	}
	return paysheet.Weights(payoutRatesByAdmin)
}

func pickUniformly(videoCount int) (int, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(videoCount)))
	if err != nil {
		return 0, err
	}
	return int(index.Int64()), nil
}
