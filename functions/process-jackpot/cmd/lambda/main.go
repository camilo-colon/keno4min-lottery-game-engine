package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/adapters"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/config"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/database"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/handler"
)

// EventInput recibe el game recién sorteado por DrawBalls (con sus balotas).
type EventInput struct {
	Game domain.Game `json:"game"`
}

type Output struct {
	GameID         string `json:"gameId"`
	ProcessedClubs int    `json:"processedClubs"`
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
	clubStore := adapters.NewClubStore(db.DB)
	runStore := adapters.NewRunStore(db.DB)
	txManager := adapters.NewMongoTransactionManager(db.Client)
	randomizer := adapters.NewCryptoRandomizer()

	handler := handler.NewProcessJackpotHandler(ticketStore, clubStore, runStore, txManager, randomizer)
	processed, err := handler.Handle(ctx, input.Game)
	if err != nil {
		return nil, fmt.Errorf("error processing jackpot: %w", err)
	}

	return &Output{GameID: input.Game.ID, ProcessedClubs: processed}, nil
}

func main() {
	lambda.Start(lambdaHandler)
}
