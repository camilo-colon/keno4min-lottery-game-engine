package domain_test

import (
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

func TestNewJackpotAward(t *testing.T) {
	winner := domain.Ticket{
		ID:     "t1",
		Cupon:  "ABC123",
		Round:  42,
		GameID: "g1",
		ClubID: "c1",
	}

	award := domain.NewJackpotAward(winner, 25_000_000)

	if award.ClubID != "c1" {
		t.Errorf("ClubID = %q, want c1", award.ClubID)
	}
	if award.GameID != "g1" {
		t.Errorf("GameID = %q, want g1", award.GameID)
	}
	if award.TicketID != "t1" {
		t.Errorf("TicketID = %q, want t1", award.TicketID)
	}
	if award.Cupon != "ABC123" {
		t.Errorf("Cupon = %q, want ABC123", award.Cupon)
	}
	if award.Round != 42 {
		t.Errorf("Round = %d, want 42", award.Round)
	}
	if award.Value != 25_000_000 {
		t.Errorf("Value = %d, want 25000000", award.Value)
	}
	// Es una entidad: debe generar identidad y timestamp.
	if award.ID == "" {
		t.Error("ID no debe estar vacío (entidad con identidad propia)")
	}
	if award.CreatedAt.IsZero() {
		t.Error("CreatedAt no debe estar en cero")
	}
}
