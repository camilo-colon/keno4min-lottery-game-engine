package usecases

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	draw "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/draw/domain"
	video "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/domain"
)

type drawRepository struct {
	update func(context.Context, *draw.Draw) error
}

func (r drawRepository) Update(ctx context.Context, d *draw.Draw) error {
	return r.update(ctx, d)
}

type videoRepository struct {
	getAll func(context.Context) ([]video.Video, error)
}

func (r videoRepository) GetAll(ctx context.Context) ([]video.Video, error) {
	return r.getAll(ctx)
}

func TestDrawVideoSelectsAndPersistsMatchingVideo(t *testing.T) {
	for _, videos := range [][]video.Video{
		{{ID: "0007", Balls: []int{63, 80, 37}}},
		{{ID: "0007", Balls: []int{63, 80, 37}}, {ID: "3573", Balls: []int{31, 68, 57}}},
	} {
		ctx := t.Context()
		input := draw.New("draw-1", 42, time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC))
		original := *input
		var selected video.Video
		var persisted *draw.Draw
		reads, writes := 0, 0
		usecase := NewDrawVideo(
			drawRepository{update: func(gotCtx context.Context, d *draw.Draw) error {
				writes++
				if gotCtx != ctx || d == input || d.ID != input.ID || d.TIR != input.TIR ||
					d.Status != draw.StatusDrawn || d.StartsAt != input.StartsAt || d.ClosesAt != input.ClosesAt || reads != 1 {
					t.Fatalf("unexpected update: id=%s, reads=%d", d.ID, reads)
				}
				selected = video.Video{ID: d.VideoID, Balls: d.Balls}
				persisted = d
				return nil
			}},
			videoRepository{getAll: func(gotCtx context.Context) ([]video.Video, error) {
				reads++
				if gotCtx != ctx {
					t.Fatal("context was not forwarded")
				}
				return videos, nil
			}},
		)
		result, err := usecase.Execute(ctx, input)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, v := range videos {
			if reflect.DeepEqual(selected, v) {
				found = true
			}
		}
		if !found || writes != 1 || reads != 1 {
			t.Fatalf("selected=%+v, reads=%d, writes=%d", selected, reads, writes)
		}
		want := original
		want.VideoID, want.Balls, want.Status = selected.ID, selected.Balls, draw.StatusDrawn
		if result != persisted || !reflect.DeepEqual(result, &want) || !reflect.DeepEqual(input, &original) {
			t.Fatalf("result=%+v, input=%+v", result, input)
		}
		result.Balls[0] = 999
		for _, v := range videos {
			if v.Balls[0] == 999 {
				t.Fatal("result aliases catalog balls")
			}
		}
	}
}

func TestDrawVideoDoesNotUpdateWhenCatalogFailsOrIsEmpty(t *testing.T) {
	failure := errors.New("catalog unavailable")
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{name: "query error", err: failure, want: failure},
		{name: "no videos", want: ErrNoVideos},
	} {
		t.Run(tc.name, func(t *testing.T) {
			usecase := NewDrawVideo(
				drawRepository{update: func(context.Context, *draw.Draw) error {
					t.Fatal("must not update without a selected video")
					return nil
				}},
				videoRepository{getAll: func(context.Context) ([]video.Video, error) {
					return nil, tc.err
				}},
			)
			result, err := usecase.Execute(t.Context(), &draw.Draw{ID: "draw-1"})
			if result != nil || !errors.Is(err, tc.want) {
				t.Fatalf("Execute() = %+v, %v", result, err)
			}
		})
	}
}

func TestDrawVideoPreservesInputWhenUpdateFails(t *testing.T) {
	failure := errors.New("update failed")
	input := &draw.Draw{ID: "draw-1", VideoID: "old", Balls: []int{1, 2}}
	want := &draw.Draw{ID: "draw-1", VideoID: "old", Balls: []int{1, 2}}
	usecase := NewDrawVideo(
		drawRepository{update: func(context.Context, *draw.Draw) error { return failure }},
		videoRepository{getAll: func(context.Context) ([]video.Video, error) {
			return []video.Video{{ID: "3573", Balls: []int{63, 80}}}, nil
		}},
	)
	result, err := usecase.Execute(t.Context(), input)
	if result != nil || !errors.Is(err, failure) || !reflect.DeepEqual(input, want) {
		t.Fatalf("Execute() = %+v, %v; input=%+v", result, err, input)
	}
}
