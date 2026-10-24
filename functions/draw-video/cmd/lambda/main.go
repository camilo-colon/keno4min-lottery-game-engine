package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	drawstore "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/adapters/dynamodb"
	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/usecases"
	videostore "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/adapters/dynamodb"
)

func main() {
	drawVideo, err := setup()
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(func(ctx context.Context, input draw.Draw) (*draw.Draw, error) {
		result, err := drawVideo.Execute(ctx, &input)
		if err != nil {
			slog.Error("draw video failed", "id", input.ID, "error", err)
			return nil, err
		}
		slog.Info("draw video ready", "id", result.ID, "tir", result.TIR, "videoId", result.VideoID)
		return result, nil
	})
}

func setup() (*usecases.DrawVideo, error) {
	tableName := strings.TrimSpace(os.Getenv("DYNAMODB_TABLE"))
	if tableName == "" {
		return nil, errors.New("DYNAMODB_TABLE is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := dynamodb.NewFromConfig(cfg)
	return usecases.NewDrawVideo(
		drawstore.NewRepository(client, tableName),
		videostore.NewRepository(client, tableName),
	), nil
}
