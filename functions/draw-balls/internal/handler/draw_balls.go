package handler

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
)

// DrawBallsHandler orquesta el proceso de selección y sorteo de balotas
type DrawBallsHandler struct {
	gameStore GameRepository
	drawStore DrawRepository
}

// NewDrawBallsHandler crea un nuevo handler para ejecutar el sorteo del juego.
func NewDrawBallsHandler(gameStore GameRepository, drawStore DrawRepository) *DrawBallsHandler {
	return &DrawBallsHandler{
		gameStore: gameStore,
		drawStore: drawStore,
	}
}

// Handle ejecuta el proceso completo de sorteo de balotas
func (h *DrawBallsHandler) Handle(ctx context.Context, gameID string) (*domain.Game, error) {
	game, err := h.gameStore.FindByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	randomDraw, err := h.drawStore.GetRandomKeno4MinDraw(ctx)
	if err != nil {
		return nil, err
	}

	return h.finalizeGame(ctx, game, randomDraw)
}

// finalizeGame actualiza el juego con el draw seleccionado
func (h *DrawBallsHandler) finalizeGame(ctx context.Context, game *domain.Game, draw *domain.Draws) (*domain.Game, error) {
	balls, err := draw.ToGameBalls()
	if err != nil {
		return nil, err
	}

	game.DrawBalls(draw.Idv, *balls)

	if err := h.gameStore.UpdateGame(ctx, game); err != nil {
		return nil, err
	}

	return game, nil
}
