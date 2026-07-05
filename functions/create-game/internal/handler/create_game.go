package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/config"
	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/database"
	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/repository"
)

type CreateGameHandler struct {
	counterRepo *repository.CounterRepository
	gameRepo    *repository.GameRepository
}

func NewCreateGameHandler(counterRepo *repository.CounterRepository, gameRepo *repository.GameRepository) *CreateGameHandler {
	return &CreateGameHandler{
		counterRepo: counterRepo,
		gameRepo:    gameRepo,
	}
}

// Handle ejecuta la lógica de creación de juego
func (h *CreateGameHandler) Handle(ctx context.Context) (*domain.Game, error) {
	// Obtener el siguiente round
	round, err := h.counterRepo.IncrementAndGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("error incrementing round counter: %w", err)
	}

	// Crear el juego
	game := domain.NewGame(round)

	// Guardar en MongoDB
	if err := h.gameRepo.Create(ctx, game); err != nil {
		return nil, fmt.Errorf("error creating game: %w", err)
	}

	log.Printf("Game created successfully: ID=%s, Round=%d", game.ID, game.Round)
	return game, nil
}

// Setup inicializa las dependencias del handler
func Setup(ctx context.Context) (*CreateGameHandler, func(), error) {
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
	counterRepo := repository.NewCounterRepository(db.DB)
	gameRepo := repository.NewGameRepository(db.DB)

	// Crear handler
	handler := NewCreateGameHandler(counterRepo, gameRepo)

	return handler, cleanup, nil
}
