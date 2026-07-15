package handler_test

import (
	"context"

	"github.com/cronos/keno4min-lottery-game-engine/functions/process-jackpot/internal/domain"
)

// fakeTx ejecuta la función directamente (sin retry ni rollback real): suficiente
// para verificar la orquestación y la propagación de errores del handler.
type fakeTx struct{}

func (fakeTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeRuns simula el guardián de idempotencia.
type fakeRuns struct {
	err       error
	markCalls int
}

func (f *fakeRuns) Mark(ctx context.Context, gameID, clubID string) error {
	f.markCalls++
	return f.err
}

// fakeClubs registra llamadas al repositorio de clubes.
type fakeClubs struct {
	club            *domain.Club
	findErr         error
	findCalls       int
	incrementResult *domain.Jackpot
	incrementErr    error
	incrementCalls  int
	incrementedWith int64
	resetErr        error
	resetCalls      int
	resetWith       domain.Jackpot
}

func (f *fakeClubs) FindByID(ctx context.Context, clubID string) (*domain.Club, error) {
	f.findCalls++
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.club, nil
}

func (f *fakeClubs) IncrementJackpot(ctx context.Context, clubID string, amount int64) (*domain.Jackpot, error) {
	f.incrementCalls++
	f.incrementedWith = amount
	if f.incrementErr != nil {
		return nil, f.incrementErr
	}
	return f.incrementResult, nil
}

func (f *fakeClubs) ResetJackpot(ctx context.Context, clubID string, jackpot domain.Jackpot) error {
	f.resetCalls++
	f.resetWith = jackpot
	return f.resetErr
}

// fakeTickets registra llamadas al repositorio de tickets.
type fakeTickets struct {
	clubIDs        []string
	clubIDsErr     error
	tickets        []domain.Ticket
	ticketsErr     error
	assignErr      error
	assignCalls    int
	assignedTo     string
	assignedAmount int64
}

func (f *fakeTickets) FindClubIDsByGame(ctx context.Context, gameID string) ([]string, error) {
	return f.clubIDs, f.clubIDsErr
}

func (f *fakeTickets) FindByClubAndGame(ctx context.Context, clubID, gameID string) ([]domain.Ticket, error) {
	return f.tickets, f.ticketsErr
}

func (f *fakeTickets) AssignJackpot(ctx context.Context, ticketID string, amount int64) error {
	f.assignCalls++
	f.assignedTo = ticketID
	f.assignedAmount = amount
	return f.assignErr
}

// fakeAwards registra los premios de jackpot históricos.
type fakeAwards struct {
	recordCalls int
	recorded    domain.JackpotAward
	err         error
}

func (f *fakeAwards) Record(ctx context.Context, award domain.JackpotAward) error {
	f.recordCalls++
	f.recorded = award
	return f.err
}

// fakeRng devuelve valores deterministas.
type fakeRng struct {
	intn    int
	between int64
}

func (f fakeRng) Intn(n int) int                    { return f.intn }
func (f fakeRng) Int64Between(min, max int64) int64 { return f.between }

// losingTicket es un ticket sin premio ganado (Win 0, ya persistido por
// update-tickets), por lo que su utilidad para la casa es igual a lo apostado.
func losingTicket(id string, state domain.TicketState, total int64) domain.Ticket {
	return domain.Ticket{
		ID:    id,
		State: state,
		Total: total,
	}
}

// drawnGame construye un game con balotas (mask vacía basta para los tests).
func drawnGame(id string) domain.Game {
	return domain.Game{ID: id, Balls: &domain.GameBalls{Mask: domain.Bitmask{}}}
}
