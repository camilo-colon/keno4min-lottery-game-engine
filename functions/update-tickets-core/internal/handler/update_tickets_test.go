package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-core/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-core/internal/handler"
)

// mask construye un Bitmask a partir de números (1..80), igual que en producción.
func mask(nums ...int64) domain.Bitmask {
	var b domain.Bitmask
	for _, n := range nums {
		pos := n - 1
		if pos < 64 {
			b.Mask1 |= int64(1) << uint(pos)
		} else {
			b.Mask2 |= int64(1) << uint(pos-64)
		}
	}
	return b
}

// pendingTicket arma un ticket PENDING del core con una sola apuesta.
func pendingTicket(id string, money int64, nums ...int64) domain.Ticket {
	return domain.Ticket{
		ID:    id,
		Game:  domain.GameKeno4Min,
		State: domain.PENDING,
		Total: money,
		Metadata: domain.Metadata{
			Bets: []domain.Bet{{Money: money, Nums: nums}},
		},
	}
}

func TestHandleFailsWithoutDrawnBalls(t *testing.T) {
	h := handler.NewUpdateTicketsHandler(&fakeTickets{}, 0)

	_, err := h.Handle(context.Background(), domain.Game{ID: "g1"}) // Balls == nil
	if err == nil {
		t.Fatal("se esperaba error cuando el game no tiene balotas")
	}
}

func TestHandleResolvesWinningTicket(t *testing.T) {
	balls := mask(7, 9, 11)
	winning := pendingTicket("t1", 50000, 7, 9, 11) // 3/3 aciertos → factor 5000
	tickets := &fakeTickets{pages: [][]domain.Ticket{{winning}}}

	h := handler.NewUpdateTicketsHandler(tickets, 0)
	game := drawnGame("g1", []int64{7, 9, 11}, balls)

	count, err := h.Handle(context.Background(), game)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	if len(tickets.updateCall) != 1 || len(tickets.updateCall[0]) != 1 {
		t.Fatalf("se esperaba un solo BulkWrite con 1 ticket, got %+v", tickets.updateCall)
	}
	got := tickets.updateCall[0][0]
	if got.State != domain.WINNING {
		t.Errorf("State = %q, want WINNING", got.State)
	}
	if got.Win <= 0 {
		t.Errorf("Win = %d, want > 0", got.Win)
	}
}

func TestHandleResolvesLosingTicket(t *testing.T) {
	balls := mask(7, 9, 11)
	losing := pendingTicket("t2", 50000, 1, 2, 3) // 0 aciertos
	tickets := &fakeTickets{pages: [][]domain.Ticket{{losing}}}

	h := handler.NewUpdateTicketsHandler(tickets, 0)
	game := drawnGame("g1", []int64{7, 9, 11}, balls)

	count, err := h.Handle(context.Background(), game)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	got := tickets.updateCall[0][0]
	if got.State != domain.LOSS {
		t.Errorf("State = %q, want LOSS", got.State)
	}
	if got.Win != 0 {
		t.Errorf("Win = %d, want 0", got.Win)
	}
}

func TestHandleNoPendingTicketsSkipsWrites(t *testing.T) {
	tickets := &fakeTickets{pages: [][]domain.Ticket{{}}}

	h := handler.NewUpdateTicketsHandler(tickets, 0)
	game := drawnGame("g1", []int64{7, 9, 11}, mask(7, 9, 11))

	count, err := h.Handle(context.Background(), game)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
	if len(tickets.updateCall) != 0 {
		t.Errorf("no se esperaban escrituras, got %d", len(tickets.updateCall))
	}
}

func TestHandlePaginatesAcrossBatches(t *testing.T) {
	balls := mask(7, 9, 11)
	page1 := []domain.Ticket{
		pendingTicket("t1", 100, 1),
		pendingTicket("t2", 100, 2),
	}
	page2 := []domain.Ticket{
		pendingTicket("t3", 100, 3),
	}
	tickets := &fakeTickets{pages: [][]domain.Ticket{page1, page2}}

	h := handler.NewUpdateTicketsHandler(tickets, 2)
	game := drawnGame("g1", []int64{7, 9, 11}, balls)

	count, err := h.Handle(context.Background(), game)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	if len(tickets.updateCall) != 2 {
		t.Fatalf("se esperaban 2 BulkWrite (uno por página), got %d", len(tickets.updateCall))
	}
	if tickets.findCalls != 3 { // 2 páginas con datos + 1 página vacía que corta el loop
		t.Errorf("findCalls = %d, want 3", tickets.findCalls)
	}
}

func TestHandlePropagatesFindError(t *testing.T) {
	boom := errors.New("boom")
	tickets := &fakeTickets{findErr: boom}

	h := handler.NewUpdateTicketsHandler(tickets, 0)
	game := drawnGame("g1", []int64{7, 9, 11}, mask(7, 9, 11))

	_, err := h.Handle(context.Background(), game)
	if !errors.Is(err, boom) {
		t.Fatalf("el error de lectura debe propagarse, got %v", err)
	}
}

func TestHandlePropagatesUpdateError(t *testing.T) {
	boom := errors.New("boom")
	winning := pendingTicket("t1", 100, 7)
	tickets := &fakeTickets{pages: [][]domain.Ticket{{winning}}, updateErr: boom}

	h := handler.NewUpdateTicketsHandler(tickets, 0)
	game := drawnGame("g1", []int64{7, 9, 11}, mask(7, 9, 11))

	_, err := h.Handle(context.Background(), game)
	if !errors.Is(err, boom) {
		t.Fatalf("el error de escritura debe propagarse, got %v", err)
	}
}
