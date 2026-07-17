package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/adapters"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/config"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/database"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/handler"
)

// EventInput recibe el game recién sorteado por DrawBalls (con sus balotas).
type EventInput struct {
	Game domain.Game `json:"game"`
}

type Output struct {
	GameID          string `json:"gameId"`
	ResolvedTickets int    `json:"resolvedTickets"`
}

var (
	dbMu     sync.Mutex
	cachedDB *database.MongoDB
)

// getDB reutiliza la conexión a MongoDB entre invocaciones tibias: el contenedor
// Lambda sobrevive al handler, así que reconectar en cada invocación pagaría el
// handshake de nuevo y rotaría slots de conexión en Atlas. Si la conexión falla
// no se cachea nada, y la siguiente invocación vuelve a intentarlo.
func getDB(ctx context.Context, cfg *config.Config) (*database.MongoDB, error) {
	dbMu.Lock()
	defer dbMu.Unlock()

	if cachedDB != nil {
		return cachedDB, nil
	}

	db, err := database.Connect(ctx, cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		return nil, err
	}

	cachedDB = db
	return cachedDB, nil
}

func lambdaHandler(ctx context.Context, input EventInput) (*Output, error) {
	cfg, err := config.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	// Conectar a MongoDB
	db, err := getDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

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
