package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/domain"
	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets/internal/handler"
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

func TestHandleFailsWithoutDrawnBalls(t *testing.T) {
	h := handler.NewUpdateTicketsHandler(&fakeTickets{}, 0)

	_, err := h.Handle(context.Background(), domain.Game{ID: "g1"}) // Balls == nil
	if err == nil {
		t.Fatal("se esperaba error cuando el game no tiene balotas")
	}
}

func TestHandleResolvesWinningTicket(t *testing.T) {
	balls := mask(7, 9, 11)
	winning := domain.Ticket{
		ID:    "t1",
		State: domain.DRAWING,
		Total: 50000,
		Bets:  []domain.Bet{{Money: 50000, Bitmask: mask(7, 9, 11)}}, // 3/3 aciertos → factor 5000
	}
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
	if len(got.Balls) != 3 || got.Balls[0] != 7 || got.Balls[1] != 9 || got.Balls[2] != 11 {
		t.Errorf("Balls no se copiaron del juego: got %v", got.Balls)
	}
}

func TestHandleResolvesLosingTicket(t *testing.T) {
	balls := mask(7, 9, 11)
	losing := domain.Ticket{
		ID:    "t2",
		State: domain.DRAWING,
		Total: 50000,
		Bets:  []domain.Bet{{Money: 50000, Bitmask: mask(1, 2, 3)}}, // 0 aciertos
	}
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

func TestHandleNoDrawingTicketsSkipsWrites(t *testing.T) {
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
		{ID: "t1", State: domain.DRAWING, Bets: []domain.Bet{{Money: 100, Bitmask: mask(1)}}},
		{ID: "t2", State: domain.DRAWING, Bets: []domain.Bet{{Money: 100, Bitmask: mask(2)}}},
	}
	page2 := []domain.Ticket{
		{ID: "t3", State: domain.DRAWING, Bets: []domain.Bet{{Money: 100, Bitmask: mask(3)}}},
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
	winning := domain.Ticket{ID: "t1", State: domain.DRAWING, Bets: []domain.Bet{{Money: 100, Bitmask: mask(7)}}}
	tickets := &fakeTickets{pages: [][]domain.Ticket{{winning}}, updateErr: boom}

	h := handler.NewUpdateTicketsHandler(tickets, 0)
	game := drawnGame("g1", []int64{7, 9, 11}, mask(7, 9, 11))

	_, err := h.Handle(context.Background(), game)
	if !errors.Is(err, boom) {
		t.Fatalf("el error de escritura debe propagarse, got %v", err)
	}
}
