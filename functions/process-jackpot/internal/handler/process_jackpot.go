package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/ports"
)

// ProcessJackpotHandler orquesta el incremento (y eventual juego) del jackpot
// de cada club que participó en un juego.
type ProcessJackpotHandler struct {
	tickets ports.TicketRepository
	clubs   ports.ClubRepository
	runs    ports.RunRepository
	tx      ports.TransactionManager
	rng     ports.Randomizer
}

// NewProcessJackpotHandler crea un nuevo handler para procesar el jackpot.
func NewProcessJackpotHandler(
	tickets ports.TicketRepository,
	clubs ports.ClubRepository,
	runs ports.RunRepository,
	tx ports.TransactionManager,
	rng ports.Randomizer,
) *ProcessJackpotHandler {
	return &ProcessJackpotHandler{
		tickets: tickets,
		clubs:   clubs,
		runs:    runs,
		tx:      tx,
		rng:     rng,
	}
}

// Handle procesa el jackpot de todos los clubes que participaron en el juego.
// El sorteo (balotas) viene en game.Balls, fuente de la verdad del juego recién
// ejecutado por DrawBalls. Devuelve la cantidad de clubes procesados.
func (h *ProcessJackpotHandler) Handle(ctx context.Context, game domain.Game) (int, error) {
	if game.Balls == nil {
		return 0, fmt.Errorf("game %s has no drawn balls", game.ID)
	}
	balls := game.Balls.Mask

	clubIDs, err := h.tickets.FindClubIDsByGame(ctx, game.ID)
	if err != nil {
		return 0, err
	}

	for _, clubID := range clubIDs {
		if err := h.processClub(ctx, game.ID, clubID, balls); err != nil {
			return 0, err
		}
	}

	return len(clubIDs), nil
}

// processClub ejecuta, de forma atómica e idempotente, el incremento y eventual
// juego del jackpot de un club. Toda la unidad de trabajo —marca de idempotencia,
// lecturas y escrituras de dinero— vive en una sola transacción: o se commitea
// completa, o se hace rollback de todo.
func (h *ProcessJackpotHandler) processClub(ctx context.Context, gameID, clubID string, balls domain.Bitmask) error {
	err := h.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Guardián de idempotencia: si ya se procesó, aborta la transacción.
		if err := h.runs.Mark(txCtx, gameID, clubID); err != nil {
			return err
		}

		club, err := h.clubs.FindByID(txCtx, clubID)
		if err != nil {
			return err
		}

		tickets, err := h.tickets.FindByClubAndGame(txCtx, clubID, gameID)
		if err != nil {
			return err
		}

		increment := domain.JackpotIncrement(club.JP1, tickets, balls)

		// Si no hay utilidad positiva el pozo no cambia: evaluamos ShouldPlay
		// sobre el jackpot actual sin escribir.
		jackpot := &club.JP1
		if increment > 0 {
			jackpot, err = h.clubs.IncrementJackpot(txCtx, clubID, increment)
			if err != nil {
				return err
			}
		}

		if jackpot.ShouldPlay() {
			return h.playJackpot(txCtx, clubID, club.Config.Jackpot, *jackpot, tickets)
		}
		return nil
	})

	// Ya procesado en una corrida previa: skip idempotente, no es un error.
	if errors.Is(err, ports.ErrAlreadyProcessed) {
		return nil
	}
	return err
}

// playJackpot sortea un ticket ganador, le asigna el pozo acumulado y resetea el
// jp1 del club con la config. Los cancelados ya vienen filtrados por el puerto de
// tickets, así que cualquier ticket recibido es elegible.
func (h *ProcessJackpotHandler) playJackpot(ctx context.Context, clubID string, cfg domain.JackpotConfig, jackpot domain.Jackpot, tickets []domain.Ticket) error {
	if len(tickets) == 0 {
		// Sin tickets no hay a quién pagar: el pozo se mantiene.
		return nil
	}
	winner := tickets[h.rng.Intn(len(tickets))]

	if err := h.tickets.AssignJackpot(ctx, winner.ID, jackpot.Value); err != nil {
		return err
	}

	target := h.rng.Int64Between(cfg.Min, cfg.Max)
	fresh := domain.NewJackpotFromConfig(cfg, target)
	return h.clubs.ResetJackpot(ctx, clubID, fresh)
}
