package draw

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
	"uuid"
)

type memoryRepository struct {
	existing   *Draw
	findErr    error
	counterErr error
	saveErr    error
	counter    uint64
	saves      int
	requestID  string
}

func (r *memoryRepository) FindByRequestID(ctx context.Context, id string) (*Draw, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.requestID = id
	return r.existing, r.findErr
}
func (r *memoryRepository) NextTIR(context.Context) (uint64, error) {
	if r.counterErr != nil {
		return 0, r.counterErr
	}
	r.counter++
	return r.counter, nil
}
func (r *memoryRepository) Save(_ context.Context, id string, d *Draw) error {
	r.saves++
	r.requestID = id
	if r.saveErr != nil {
		return r.saveErr
	}
	r.existing = d
	return nil
}

func TestCreateAndRepeatRequest(t *testing.T) {
	r := &memoryRepository{counter: 10}
	create := NewCreate(r)
	before := time.Now().Add(-time.Millisecond)
	result, err := create.Execute(context.Background(), "request-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(result.ID); err != nil {
		t.Fatal(err)
	}
	if result.TIR != 11 || result.Status != "open" || result.ClosesAt.Sub(result.StartsAt) != Duration {
		t.Fatalf("unexpected draw: %+v", result)
	}
	if result.StartsAt.Before(before) || result.StartsAt.After(time.Now()) || result.StartsAt.Location() != time.UTC {
		t.Fatalf("unexpected startsAt: %s", result.StartsAt)
	}
	again, err := create.Execute(context.Background(), "request-1")
	if err != nil || again != result || r.counter != 11 || r.saves != 1 || r.requestID != "request-1" {
		t.Fatalf("duplicate consumed a TIR or changed the draw: %+v, %v", r, err)
	}
}

func TestCreateRejectsBlankRequestBeforeAllocating(t *testing.T) {
	for _, id := range []string{"", " \t\n"} {
		r := &memoryRepository{}
		_, err := NewCreate(r).Execute(context.Background(), id)
		if err == nil || r.counter != 0 || r.saves != 0 {
			t.Fatalf("unexpected result for %q: %+v, %v", id, r, err)
		}
	}
}

func TestCreatePropagatesErrorsAndStops(t *testing.T) {
	failure := errors.New("unavailable")
	conflict := &DrawConflictError{Cause: failure}
	for _, tc := range []struct {
		name       string
		repository memoryRepository
		want       error
		counter    uint64
		saves      int
	}{
		{"lookup", memoryRepository{findErr: failure}, failure, 0, 0},
		{"counter", memoryRepository{counterErr: failure}, failure, 0, 0},
		{"save", memoryRepository{saveErr: failure}, failure, 1, 1},
		{"conflict", memoryRepository{saveErr: conflict}, conflict, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.repository
			result, err := NewCreate(&r).Execute(context.Background(), "request")
			if result != nil || err != tc.want || r.counter != tc.counter || r.saves != tc.saves {
				t.Fatalf("unexpected result: %+v, %v, repository=%+v", result, err, r)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewCreate(&memoryRepository{}).Execute(ctx, "request")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("context was lost: %v", err)
	}
}

func TestDrawJSONContract(t *testing.T) {
	start := time.Date(2026, 10, 21, 10, 0, 0, 123456789, time.FixedZone("Bogota", -5*60*60))
	d := New("draw-1", 42, start)
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":"draw-1","tir":42,"status":"open","startsAt":"2026-10-21T15:00:00.123456789Z","closesAt":"2026-10-21T15:02:37.123456789Z"}`
	if string(data) != want {
		t.Fatalf("unexpected response: %s", data)
	}
}
