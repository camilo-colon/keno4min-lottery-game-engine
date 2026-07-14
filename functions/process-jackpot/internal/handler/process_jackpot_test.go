package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/handler"
	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/ports"
)

var jackpotConfig = domain.JackpotConfig{
	Min:     15_000_000,
	Max:     25_000_000,
	Percent: 3,
	Seed:    5_000_000,
}

func newClub() *domain.Club {
	return &domain.Club{
		ID:     "c1",
		JP1:    domain.Jackpot{Percent: 1, Value: 100, Target: 999_999_999},
		Config: domain.ClubConfig{Jackpot: jackpotConfig},
	}
}

func TestHandleFailsWithoutDrawnBalls(t *testing.T) {
	h := handler.NewProcessJackpotHandler(&fakeTickets{}, &fakeClubs{}, &fakeRuns{}, &fakeAwards{}, fakeTx{}, fakeRng{})

	_, err := h.Handle(context.Background(), domain.Game{ID: "g1"}) // Balls == nil
	if err == nil {
		t.Fatal("se esperaba error cuando el game no tiene balotas")
	}
}

func TestProcessClubSkipsWhenAlreadyProcessed(t *testing.T) {
	tickets := &fakeTickets{clubIDs: []string{"c1"}}
	clubs := &fakeClubs{club: newClub()}
	runs := &fakeRuns{err: ports.ErrAlreadyProcessed}

	h := handler.NewProcessJackpotHandler(tickets, clubs, runs, &fakeAwards{}, fakeTx{}, fakeRng{})

	processed, err := h.Handle(context.Background(), drawnGame("g1"))
	if err != nil {
		t.Fatalf("skip idempotente no debe ser error: %v", err)
	}
	if processed != 1 {
		t.Errorf("processed = %d, want 1", processed)
	}
	if runs.markCalls != 1 {
		t.Errorf("markCalls = %d, want 1", runs.markCalls)
	}
	// No debe leer ni escribir nada tras el skip.
	if clubs.findCalls != 0 {
		t.Errorf("FindByID no debía llamarse tras el skip, calls = %d", clubs.findCalls)
	}
	if clubs.incrementCalls != 0 || tickets.assignCalls != 0 || clubs.resetCalls != 0 {
		t.Errorf("no debía haber escrituras tras el skip")
	}
}

func TestProcessClubIncrementsWithoutPlaying(t *testing.T) {
	tickets := &fakeTickets{
		clubIDs: []string{"c1"},
		tickets: []domain.Ticket{losingTicket("t1", domain.PAYED, 1_000_000)}, // utilidad +1M
	}
	clubs := &fakeClubs{
		club:            newClub(),
		incrementResult: &domain.Jackpot{Value: 16_000_000, Target: 20_000_000}, // < target → no juega
	}

	h := handler.NewProcessJackpotHandler(tickets, clubs, &fakeRuns{}, &fakeAwards{}, fakeTx{}, fakeRng{})

	if _, err := h.Handle(context.Background(), drawnGame("g1")); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if clubs.incrementCalls != 1 {
		t.Errorf("incrementCalls = %d, want 1", clubs.incrementCalls)
	}
	if clubs.incrementedWith != 10_000 { // 1_000_000 * 1 / 100
		t.Errorf("incrementedWith = %d, want 10000", clubs.incrementedWith)
	}
	if tickets.assignCalls != 0 || clubs.resetCalls != 0 {
		t.Errorf("no debía jugar el jackpot: assign=%d reset=%d", tickets.assignCalls, clubs.resetCalls)
	}
}

func TestProcessClubPlaysJackpot(t *testing.T) {
	winner := losingTicket("t1", domain.PAYED, 1_000_000) // utilidad +1M
	winner.Cupon = "ABC123"
	winner.Round = 42
	winner.GameID = "g1"
	winner.ClubID = "c1"

	tickets := &fakeTickets{
		clubIDs: []string{"c1"},
		// El puerto ya filtra cancelados: aquí solo llegan tickets elegibles.
		tickets: []domain.Ticket{winner},
	}
	clubs := &fakeClubs{
		club:            newClub(),
		incrementResult: &domain.Jackpot{Value: 25_000_000, Target: 20_000_000}, // >= target → juega
	}
	awards := &fakeAwards{}
	rng := fakeRng{intn: 0, between: 20_000_000}

	h := handler.NewProcessJackpotHandler(tickets, clubs, &fakeRuns{}, awards, fakeTx{}, rng)

	if _, err := h.Handle(context.Background(), drawnGame("g1")); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	// Asigna el pozo al ticket ganador.
	if tickets.assignCalls != 1 {
		t.Fatalf("assignCalls = %d, want 1", tickets.assignCalls)
	}
	if tickets.assignedTo != "t1" {
		t.Errorf("ganador = %q, want t1", tickets.assignedTo)
	}
	if tickets.assignedAmount != 25_000_000 {
		t.Errorf("premio asignado = %d, want 25000000 (jackpot.Value)", tickets.assignedAmount)
	}

	// Registra el premio en el histórico con los datos del ganador.
	if awards.recordCalls != 1 {
		t.Fatalf("recordCalls = %d, want 1", awards.recordCalls)
	}
	got := awards.recorded
	if got.ClubID != "c1" || got.GameID != "g1" || got.TicketID != "t1" ||
		got.Cupon != "ABC123" || got.Round != 42 || got.Value != 25_000_000 {
		t.Errorf("award registrado = %+v, want club=c1 game=g1 ticket=t1 cupon=ABC123 round=42 value=25000000", got)
	}

	// Resetea el jp1 con la config y el target aleatorio.
	if clubs.resetCalls != 1 {
		t.Fatalf("resetCalls = %d, want 1", clubs.resetCalls)
	}
	want := domain.Jackpot{Percent: 3, Target: 20_000_000, Value: 5_000_000, Min: 15_000_000, Max: 25_000_000}
	if clubs.resetWith != want {
		t.Errorf("resetWith = %+v, want %+v", clubs.resetWith, want)
	}
}

func TestProcessClubZeroIncrementSkipsWrites(t *testing.T) {
	tickets := &fakeTickets{clubIDs: []string{"c1"}, tickets: nil} // sin tickets → incremento 0
	clubs := &fakeClubs{club: newClub()}                           // jp1.Value 100 < Target → no juega
	runs := &fakeRuns{}

	h := handler.NewProcessJackpotHandler(tickets, clubs, runs, &fakeAwards{}, fakeTx{}, fakeRng{})

	if _, err := h.Handle(context.Background(), drawnGame("g1")); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if runs.markCalls != 1 {
		t.Errorf("markCalls = %d, want 1 (la marca siempre se registra)", runs.markCalls)
	}
	if clubs.incrementCalls != 0 {
		t.Errorf("incremento 0 no debe escribir: incrementCalls = %d", clubs.incrementCalls)
	}
	if tickets.assignCalls != 0 || clubs.resetCalls != 0 {
		t.Errorf("no debía jugar el jackpot")
	}
}

func TestProcessClubPropagatesWriteError(t *testing.T) {
	tickets := &fakeTickets{
		clubIDs: []string{"c1"},
		tickets: []domain.Ticket{losingTicket("t1", domain.PAYED, 1_000_000)},
	}
	boom := errors.New("boom")
	clubs := &fakeClubs{
		club:            newClub(),
		incrementResult: &domain.Jackpot{Value: 25_000_000, Target: 20_000_000},
		resetErr:        boom, // falla el reset dentro de la transacción
	}
	rng := fakeRng{intn: 0, between: 20_000_000}

	h := handler.NewProcessJackpotHandler(tickets, clubs, &fakeRuns{}, &fakeAwards{}, fakeTx{}, rng)

	_, err := h.Handle(context.Background(), drawnGame("g1"))
	if !errors.Is(err, boom) {
		t.Fatalf("el error de escritura debe propagarse, got %v", err)
	}
}

func TestHandleProcessesAllClubs(t *testing.T) {
	tickets := &fakeTickets{clubIDs: []string{"c1", "c2", "c3"}}
	clubs := &fakeClubs{club: newClub()}
	runs := &fakeRuns{}

	h := handler.NewProcessJackpotHandler(tickets, clubs, runs, &fakeAwards{}, fakeTx{}, fakeRng{})

	processed, err := h.Handle(context.Background(), drawnGame("g1"))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if processed != 3 {
		t.Errorf("processed = %d, want 3", processed)
	}
	if runs.markCalls != 3 {
		t.Errorf("markCalls = %d, want 3 (uno por club)", runs.markCalls)
	}
}
