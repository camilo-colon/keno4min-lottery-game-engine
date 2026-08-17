package handler

import (
	"context"
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
)

type gameRepositoryStub struct {
	game    *domain.Game
	updated *domain.Game
}

func (s *gameRepositoryStub) FindByID(context.Context, string) (*domain.Game, error) {
	return s.game, nil
}

func (s *gameRepositoryStub) UpdateGame(_ context.Context, game *domain.Game) error {
	s.updated = game
	return nil
}

func (s *gameRepositoryStub) GetHistoryGame(context.Context, int64) ([]domain.Game, error) {
	return nil, nil
}

type drawRepositoryStub struct {
	draw  *domain.Draws
	calls int
}

func (s *drawRepositoryStub) GetRandomKeno4MinDraw(context.Context) (*domain.Draws, error) {
	s.calls++
	return s.draw, nil
}

func TestHandleUsesFirstRandomDrawWithoutRTPFiltering(t *testing.T) {
	gameStore := &gameRepositoryStub{game: &domain.Game{ID: "game-1", Status: domain.BETTING}}
	drawStore := &drawRepositoryStub{draw: &domain.Draws{
		Idv: "123_1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20",
	}}
	handler := NewDrawBallsHandler(gameStore, drawStore)

	game, err := handler.Handle(context.Background(), "game-1")
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	if drawStore.calls != 1 {
		t.Fatalf("GetRandomKeno4MinDraw() calls = %d, want 1", drawStore.calls)
	}
	if gameStore.updated != game {
		t.Fatal("UpdateGame() did not receive the returned game")
	}
	if game.Status != domain.DRAWN {
		t.Fatalf("game status = %q, want %q", game.Status, domain.DRAWN)
	}
	if game.Idv != drawStore.draw.Idv {
		t.Fatalf("game idv = %q, want %q", game.Idv, drawStore.draw.Idv)
	}
	if game.Balls == nil || len(game.Balls.Nums) != 20 {
		t.Fatalf("game balls = %#v, want 20 numbers", game.Balls)
	}
}
