package usecases

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"

	bet "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bet/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/bitmask"
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

type betRepository struct {
	getByTIR func(context.Context, uint64) ([]bet.AdminBets, error)
}

func (r betRepository) GetByTIR(ctx context.Context, tir uint64) ([]bet.AdminBets, error) {
	return r.getByTIR(ctx, tir)
}

var roundWithoutBets = betRepository{getByTIR: func(context.Context, uint64) ([]bet.AdminBets, error) {
	return nil, nil
}}

var fourVideos = []video.Video{
	{ID: "three-hits", Balls: []int{1, 2, 3}},
	{ID: "two-hits-a", Balls: []int{1, 2, 60}},
	{ID: "two-hits-b", Balls: []int{1, 3, 61}},
	{ID: "two-hits-c", Balls: []int{2, 3, 62}},
}

var fiveNumbers = bet.Bet{Numbers: bitmask.MaskOf([]int{1, 2, 3, 4, 5}), StakeCents: 100}

func drawVideoWith(bets betRepository) *DrawVideo {
	return NewDrawVideo(
		drawRepository{update: func(context.Context, *draw.Draw) error { return nil }},
		videoRepository{getAll: func(context.Context) ([]video.Video, error) { return fourVideos, nil }},
		bets,
	)
}

func timesEachVideoWasDrawn(t *testing.T, usecase *DrawVideo, rounds int) map[string]int {
	timesDrawn := make(map[string]int)
	for range rounds {
		result, err := usecase.Execute(t.Context(), draw.New("draw-1", 42, time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)))
		if err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
		timesDrawn[result.VideoID]++
	}
	return timesDrawn
}

func TestDrawVideoNeverDrawsAVideoThePaysheetLeftWithoutWeight(t *testing.T) {
	usecase := drawVideoWith(betRepository{getByTIR: func(context.Context, uint64) ([]bet.AdminBets, error) {
		return []bet.AdminBets{{AdminNodeID: "admin-1", Bets: []bet.Bet{fiveNumbers}}}, nil
	}})
	timesDrawn := timesEachVideoWasDrawn(t, usecase, 300)
	if timesDrawn["three-hits"] != 0 || len(timesDrawn) != 3 {
		t.Fatalf("times each video was drawn = %v, want three-hits never drawn and the other three drawn", timesDrawn)
	}
}

func TestDrawVideoPicksUniformlyWhenThePaysheetFails(t *testing.T) {
	tests := []struct {
		name     string
		getByTIR func(context.Context, uint64) ([]bet.AdminBets, error)
	}{
		{name: "bets cannot be read", getByTIR: func(context.Context, uint64) ([]bet.AdminBets, error) {
			return nil, errors.New("query failed")
		}},
		{name: "two admins with the same bets", getByTIR: func(context.Context, uint64) ([]bet.AdminBets, error) {
			return []bet.AdminBets{
				{AdminNodeID: "admin-1", Bets: []bet.Bet{fiveNumbers}},
				{AdminNodeID: "admin-2", Bets: []bet.Bet{fiveNumbers}},
			}, nil
		}},
		{name: "reading bets panics", getByTIR: func(context.Context, uint64) ([]bet.AdminBets, error) {
			panic("unexpected failure")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timesDrawn := timesEachVideoWasDrawn(t, drawVideoWith(betRepository{getByTIR: tt.getByTIR}), 300)
			if len(timesDrawn) != 4 {
				t.Fatalf("times each video was drawn = %v, want the four videos drawn", timesDrawn)
			}
		})
	}
}

func TestDrawVideoLogsReadTimeAndComputeTimeSeparately(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	usecase := drawVideoWith(betRepository{getByTIR: func(context.Context, uint64) ([]bet.AdminBets, error) {
		return []bet.AdminBets{{AdminNodeID: "admin-1", Bets: []bet.Bet{fiveNumbers}}}, nil
	}})
	timesEachVideoWasDrawn(t, usecase, 1)
	for _, want := range []string{`msg="bets read" tir=42 admins=1 readTime=`, `msg="paysheet computed" tir=42 computeTime=`} {
		if !strings.Contains(logs.String(), want) {
			t.Fatalf("logs = %q, want them to contain %q", logs.String(), want)
		}
	}
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
		reads, writes, betReads := 0, 0, 0
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
			betRepository{getByTIR: func(gotCtx context.Context, tir uint64) ([]bet.AdminBets, error) {
				betReads++
				if gotCtx != ctx || tir != input.TIR {
					t.Fatalf("unexpected bets read: tir=%d", tir)
				}
				return nil, nil
			}},
		)
		result, err := usecase.Execute(ctx, input)
		if err != nil {
			t.Fatal(err)
		}
		if betReads != 1 {
			t.Fatalf("bets were read %d times, want 1", betReads)
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
				roundWithoutBets,
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
		roundWithoutBets,
	)
	result, err := usecase.Execute(t.Context(), input)
	if result != nil || !errors.Is(err, failure) || !reflect.DeepEqual(input, want) {
		t.Fatalf("Execute() = %+v, %v; input=%+v", result, err, input)
	}
}
