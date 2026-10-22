package draw

import (
	"context"
	"errors"
	"strings"
	"time"
	"uuid"
)

type Repository interface {
	FindByRequestID(context.Context, string) (*Draw, error)
	NextTIR(context.Context) (uint64, error)
	Save(context.Context, string, *Draw) error
}

type Create struct {
	repository Repository
}

func NewCreate(repository Repository) *Create {
	return &Create{repository: repository}
}

func (c *Create) Execute(ctx context.Context, requestID string) (*Draw, error) {
	if strings.TrimSpace(requestID) == "" {
		return nil, errors.New("requestId is required")
	}

	existing, err := c.repository.FindByRequestID(ctx, requestID)
	if err != nil || existing != nil {
		return existing, err
	}

	tir, err := c.repository.NextTIR(ctx)
	if err != nil {
		return nil, err
	}

	draw := New(uuid.NewV4().String(), tir, time.Now())
	if err := c.repository.Save(ctx, requestID, draw); err != nil {
		return nil, err
	}
	return draw, nil
}

// Lambda uses the concrete error type as errorType. Keep this name for the
// DrawConflictError retry configured in Step Functions; do not wrap it.
type DrawConflictError struct {
	Cause error
}

func (e *DrawConflictError) Error() string { return "concurrent draw creation: " + e.Cause.Error() }
func (e *DrawConflictError) Unwrap() error { return e.Cause }
