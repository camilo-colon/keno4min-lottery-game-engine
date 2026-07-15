package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/adapters"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/config"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/database"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/handler"
)

// EventInput recibe el game recién sorteado por DrawBalls (con sus balotas).
type EventInput struct {
	Game domain.Game `json:"game"`
}

type Output struct {
	GameID          string `json:"gameId"`
	ResolvedTickets int    `json:"resolvedTickets"`
}

func lambdaHandler(ctx context.Context, input EventInput) (*Output, error) {
	cfg, err := config.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	// Conectar a MongoDB
	db, err := database.Connect(ctx, cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	defer db.Disconnect(ctx)

	// Inicializar adapters
	ticketStore := adapters.NewTicketStore(db.DB)

	h := handler.NewUpdateTicketsHandler(ticketStore, handler.DefaultBatchSize)
	resolved, err := h.Handle(ctx, input.Game)
	if err != nil {
		return nil, fmt.Errorf("error updating tickets: %w", err)
	}

	return &Output{GameID: input.Game.ID, ResolvedTickets: resolved}, nil
}

func main() {
	lambda.Start(lambdaHandler)
}
