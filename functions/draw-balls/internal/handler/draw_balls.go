package handler

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/domain"
)

// DrawBallsHandler orquesta el proceso de selección y sorteo de balotas
type DrawBallsHandler struct {
	gameStore   GameRepository
	drawStore   DrawRepository
	ticketStore TicketRepository
}

// NewDrawBallsHandler crea un nuevo handler para ejecutar el sorteo del juego.
func NewDrawBallsHandler(gameStore GameRepository, drawStore DrawRepository, ticketStore TicketRepository) *DrawBallsHandler {
	return &DrawBallsHandler{
		gameStore:   gameStore,
		drawStore:   drawStore,
		ticketStore: ticketStore,
	}
}

// Handle ejecuta el proceso completo de sorteo de balotas
func (h *DrawBallsHandler) Handle(ctx context.Context, gameID string) (*domain.Game, error) {
	game, err := h.gameStore.FindByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	/*history, err := h.gameStore.GetHistoryGame(ctx, 10)
	if err != nil {
		return nil, err
	}*/

	randomDraw, err := h.drawStore.GetRandomKeno4MinDraw(ctx)

	stats, err := h.ticketStore.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	if stats.TotalIncome == 0 {
		return h.finalizeGame(ctx, game, randomDraw)
	}

	if stats.Rtp > 88.0 && stats.Rtp < 92.0 {
		return h.finalizeGame(ctx, game, randomDraw)
	}

	tickets, err := h.ticketStore.GetTicketsByGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	//decrease game algorithm
	if stats.Rtp >= 92.0 {
		maxRetries := 10
		for range maxRetries {
			currentRtp, err := calculateRTPByDraw(randomDraw, tickets)
			if err != nil {
				return nil, err
			}
			if currentRtp < stats.Rtp {
				break
			}
			randomDraw, err = h.drawStore.GetRandomKeno4MinDraw(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	//increase game algorithm
	if stats.Rtp <= 88.0 {
		maxRetries := 10
		for range maxRetries {
			currentRtp, err := calculateRTPByDraw(randomDraw, tickets)
			if err != nil {
				return nil, err
			}
			if currentRtp > stats.Rtp {
				break
			}
			randomDraw, err = h.drawStore.GetRandomKeno4MinDraw(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	return h.finalizeGame(ctx, game, randomDraw)
}

func calculateRTPByDraw(draw *domain.Draws, tickets []domain.Ticket) (float64, error) {
	var total uint
	var win uint
	gameBalls, err := draw.ToGameBalls()
	if err != nil {
		return 0, err
	}
	for _, ticket := range tickets {
		total += ticket.Total
		for _, bet := range ticket.Bets {
			nums := bet.Mask
			hits := gameBalls.Mask.Matches(&nums)
			pick := bet.Mask.Count()
			payout := (domain.PaymentTable[pick][hits] * bet.Money) / 100
			win += payout
		}
	}
	if total == 0 {
		return 0, nil
	}
	return (float64(win) / float64(total)) * 100, nil
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
