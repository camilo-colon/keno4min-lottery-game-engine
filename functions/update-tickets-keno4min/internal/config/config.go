package config

import (
	"context"
	"fmt"
	"os"

	"github.com/cronos/keno4min-lottery-game-engine/functions/update-tickets-keno4min/internal/secrets"
)

type Config struct {
	MongoURI     string
	DatabaseName string
	Stage        string
}

// Load carga la configuración desde Secrets Manager o variables de entorno
func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	cfg.Stage = getEnv("STAGE", "dev")

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
