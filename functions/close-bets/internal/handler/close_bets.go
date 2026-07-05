package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/config"
	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/database"
	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/repository"
	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/service"
)

type CloseBetsInput struct {
	GameID string `json:"gameId"`
}

type CloseBetsHandler struct {
	gameService   *service.GameService
	ticketService *service.TicketService
}

func NewCloseBetsHandler(gameService *service.GameService, ticketService *service.TicketService) *CloseBetsHandler {
	return &CloseBetsHandler{
		gameService:   gameService,
		ticketService: ticketService,
	}
}

// Handle ejecuta la lógica de cierre de apuestas
func (h *CloseBetsHandler) Handle(ctx context.Context, input CloseBetsInput) error {
	if input.GameID == "" {
		return fmt.Errorf("gameId is required")
	}

	if err := h.gameService.CloseBets(ctx, input.GameID); err != nil {
		return fmt.Errorf("error closing bets for game %s: %w", input.GameID, err)
	}

	totalUpdatedTickets, err := h.ticketService.MovePendingToDrawing(ctx, input.GameID)
	if err != nil {
		return fmt.Errorf("error updating tickets to drawing for game %s: %w", input.GameID, err)
	}

	log.Printf("Bets closed successfully for game: %s. Updated tickets to DRAWING: %d", input.GameID, totalUpdatedTickets)
	return nil
}

// Setup inicializa las dependencias del handler
func Setup(ctx context.Context) (*CloseBetsHandler, func(), error) {
	// Cargar configuración
	cfg, err := config.Load(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading config: %w", err)
	}

	// Conectar a MongoDB
	db, err := database.Connect(ctx, cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	// Función de limpieza
	cleanup := func() {
		if err := db.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}

	// Inicializar repositorios
	gameRepo := repository.NewGameRepository(db.DB)
	ticketRepo := repository.NewTicketRepository(db.DB)
	gameService := service.NewGameService(gameRepo)
	ticketService := service.NewTicketService(ticketRepo, service.DefaultBatchSize)

	// Crear handler
	handler := NewCloseBetsHandler(gameService, ticketService)

	return handler, cleanup, nil
}
