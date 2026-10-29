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
	ticketstore "github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/adapters/dynamodb"
	"github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/ticket/usecases"
)

// Draw es la salida de draw-video que Step Functions entrega a esta función.
type Draw struct {
	ID       string    `json:"id"`
	TIR      uint64    `json:"tir"`
	Status   string    `json:"status"`
	StartsAt time.Time `json:"startsAt"`
	ClosesAt time.Time `json:"closesAt"`
	Balls    []int     `json:"balls,omitempty"`
	VideoID  string    `json:"videoId,omitempty"`
}

func main() {
	updateTickets, err := setup()
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(func(ctx context.Context, input Draw) (*Draw, error) {
		updated, err := updateTickets.Execute(ctx, input.TIR, input.Balls)
		if err != nil {
			slog.Error("update tickets failed", "id", input.ID, "tir", input.TIR, "updated", updated, "error", err)
			return nil, err
		}
		slog.Info("tickets updated", "id", input.ID, "tir", input.TIR, "updated", updated)
		return &input, nil
	})
}

func setup() (*usecases.UpdateTickets, error) {
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
	return usecases.NewUpdateTickets(ticketstore.NewRepository(client, tableName)), nil
}
