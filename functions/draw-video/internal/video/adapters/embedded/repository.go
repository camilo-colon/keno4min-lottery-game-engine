package embedded

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	video "github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/domain"
	"github.com/cronos/keno4min-draw-engine/functions/draw-video/internal/video/ports"
)

const (
	ballsPerVideo = 20
	minBall       = 1
	maxBall       = 80
)

// catalogText contiene un id de video por línea. El archivo se mantiene a mano
// y se valida al iniciar la Lambda y en los tests.
//
//go:embed catalog.txt
var catalogText string

type Repository struct {
	videos []video.Video
}

var _ ports.Repository = (*Repository)(nil)

// NewRepository interpreta el catálogo incrustado una sola vez. Falla si está
// vacío o si alguna línea es inválida, para que un catálogo roto no se despliegue.
func NewRepository() (*Repository, error) {
	videos, err := parse(catalogText)
	if err != nil {
		return nil, fmt.Errorf("parse embedded catalog: %w", err)
	}
	return &Repository{videos: videos}, nil
}

// GetAll devuelve el catálogo ya interpretado, sin I/O.
func (r *Repository) GetAll(context.Context) ([]video.Video, error) {
	return r.videos, nil
}

// parse convierte el texto del catálogo en videos. Cada línea es un id con el
// formato <consecutivo>_<20 números de dos dígitos separados por comas>.
// Ignora líneas en blanco.
func parse(text string) ([]video.Video, error) {
	var videos []video.Video
	seen := make(map[string]bool)
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line) // el id es la línea recortada (tolera CRLF)
		if line == "" {
			continue
		}
		lineNumber := i + 1
		balls, err := parseBalls(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		if seen[line] {
			return nil, fmt.Errorf("line %d: duplicate id %q", lineNumber, line)
		}
		seen[line] = true
		videos = append(videos, video.Video{ID: line, Balls: balls})
	}
	if len(videos) == 0 {
		return nil, fmt.Errorf("catalog is empty")
	}
	return videos, nil
}

func parseBalls(id string) ([]int, error) {
	prefix, ballsText, found := strings.Cut(id, "_")
	if !found {
		return nil, fmt.Errorf("id %q has no '_' separator", id)
	}
	if prefix == "" || !isDigits(prefix) {
		return nil, fmt.Errorf("id %q has an empty or non-digit prefix %q", id, prefix)
	}
	parts := strings.Split(ballsText, ",")
	if len(parts) != ballsPerVideo {
		return nil, fmt.Errorf("id %q has %d balls, want %d", id, len(parts), ballsPerVideo)
	}
	balls := make([]int, 0, ballsPerVideo)
	used := make(map[int]bool)
	for _, part := range parts {
		if len(part) != 2 || !isDigits(part) {
			return nil, fmt.Errorf("id %q has ball %q that is not two digits", id, part)
		}
		ball, _ := strconv.Atoi(part)
		if ball < minBall || ball > maxBall {
			return nil, fmt.Errorf("id %q has ball %d outside %d..%d", id, ball, minBall, maxBall)
		}
		if used[ball] {
			return nil, fmt.Errorf("id %q repeats ball %d", id, ball)
		}
		used[ball] = true
		balls = append(balls, ball)
	}
	return balls, nil
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
