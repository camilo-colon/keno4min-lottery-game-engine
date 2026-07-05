package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/cronos/keno4min-lottery-game-engine/functions/draw-balls/internal/secrets"
)

type Config struct {
	MongoURI     string
	DatabaseName string
	Stage        string
	RTP          RTPConfig
}

// Load carga la configuración desde Secrets Manager o variables de entorno
func Load(ctx context.Context) (*Config, error) {
	stage := getEnv("STAGE", "dev")

	var cfg Config
	cfg.Stage = stage
	cfg.RTP = loadRTPConfig()

	// Intentar cargar desde Secrets Manager primero
	if secretName := os.Getenv("MONGODB_SECRET_NAME"); secretName != "" {
		if err := cfg.loadFromSecretsManager(ctx, secretName); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	// Fallback a variables de entorno
	cfg.MongoURI = getEnv("MONGODB_URI", "")
	cfg.DatabaseName = getEnv("DATABASE_NAME", "lottery")

	return &cfg, nil
}

func (c *Config) loadFromSecretsManager(ctx context.Context, secretName string) error {
	manager, err := secrets.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secrets manager: %w", err)
	}

	creds, err := manager.GetMongoDBCredentials(ctx, secretName)
	if err != nil {
		return fmt.Errorf("failed to get MongoDB credentials: %w", err)
	}

	c.MongoURI = creds.URI
	c.DatabaseName = creds.Database
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return defaultValue
}
