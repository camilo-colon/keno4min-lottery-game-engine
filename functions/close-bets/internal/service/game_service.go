package service

import (
	"context"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/close-bets/internal/domain"
)

type GameService struct {
	gameRepo domain.GameRepository
}

func NewGameService(gameRepo domain.GameRepository) *GameService {
	return &GameService{gameRepo: gameRepo}
}

func (s *GameService) CloseBets(ctx context.Context, gameID string) error {
	if err := s.gameRepo.UpdateStatus(ctx, gameID, domain.BETTING, domain.DRAWN); err != nil {
		return fmt.Errorf("error updating game status to DRAWN: %w", err)
	}

	return nil
}
