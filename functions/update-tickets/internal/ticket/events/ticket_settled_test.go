package events

import (
	"encoding/json"
	"testing"
	"time"
	"uuid"

	"github.com/cronos/keno4min-draw-engine/functions/update-tickets/internal/common"
)

func TestTicketSettledJSON(t *testing.T) {
	eventID := uuid.MustParse("0192f5a1-7c3e-7b2a-9d4f-1a2b3c4d5e6f")
	event := TicketSettled{
		BaseEvent:  common.BaseEvent{EventId: eventID, OccurredOn: time.Date(2026, 11, 2, 21, 30, 0, 0, time.UTC)},
		TicketID:   "0192f4e0-1a2b-7c3d-8e4f-5a6b7c8d9e0f",
		Game:       "keno4min",
		Status:     "winning",
		PrizeCents: 150000,
	}
	got, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	// detail is the contract with the Pipe and consumers, so its keys are fixed here.
	want := `{"eventId":"0192f5a1-7c3e-7b2a-9d4f-1a2b3c4d5e6f","occurredOnUtc":"2026-11-02T21:30:00Z","ticketId":"0192f4e0-1a2b-7c3d-8e4f-5a6b7c8d9e0f","game":"keno4min","status":"winning","prizeCents":150000}`
	if string(got) != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
