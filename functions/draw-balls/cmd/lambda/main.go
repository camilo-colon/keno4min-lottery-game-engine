package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/config"
	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/database"
	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/handler"
	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/repository"
)

type EventInput struct {
	GameID string `json:"gameId"`
}

func lambdaHandler(ctx context.Context, input EventInput) (*domain.Game, error) {
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

	// Inicializar repositories
	gameStore := repository.NewGameStore(db.DB)
	drawStore := repository.NewDrawStore(db.Client)
	ticketStore := repository.NewTicketStore(db.DB)

	handler := handler.NewDrawBallsHandler(gameStore, drawStore, ticketStore)
	game, err := handler.Handle(ctx, input.GameID)
	if err != nil {
		return nil, fmt.Errorf("error handling draw balls: %w", err)
	}

	return game, nil

}

func main() {
	lambda.Start(lambdaHandler)
}
