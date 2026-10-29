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
	"github.com/cronos/keno4min-draw-engine/functions/create-draw/internal/draw"
	store "github.com/cronos/keno4min-draw-engine/functions/create-draw/internal/dynamodb"
)

type event struct {
	RequestID string `json:"requestId"`
}

func main() {
	create, err := setup()
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(func(ctx context.Context, input event) (*draw.Draw, error) {
		result, err := create.Execute(ctx, input.RequestID)
		if err != nil {
			slog.Error("create draw failed", "requestId", input.RequestID, "error", err)
			return nil, err
		}
		slog.Info("draw ready", "requestId", input.RequestID, "id", result.ID, "tir", result.TIR)
		return result, nil
	})
}

func setup() (*draw.Create, error) {
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
	repository := store.NewRepository(dynamodb.NewFromConfig(cfg), tableName)
	return draw.NewCreate(repository), nil
}
