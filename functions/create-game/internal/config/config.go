package config

import (
	"context"
	"fmt"
	"os"

	"github.com/cronos/keno4min-lottery-game-engine/functions/create-game/internal/secrets"
)

type Config struct {
	MongoURI     string
	DatabaseName string
	Stage        string
}

// Load carga la configuración desde Secrets Manager o variables de entorno
func Load(ctx context.Context) (*Config, error) {
	stage := getEnv("STAGE", "dev")

	// Intentar cargar desde Secrets Manager primero
	if secretName := os.Getenv("MONGODB_SECRET_NAME"); secretName != "" {
		return loadFromSecretsManager(ctx, secretName, stage)
	}

	// Fallback a variables de entorno
	return loadFromEnv(stage), nil
}

// loadFromSecretsManager carga la configuración desde AWS Secrets Manager
func loadFromSecretsManager(ctx context.Context, secretName, stage string) (*Config, error) {
	manager, err := secrets.NewManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
	}

	creds, err := manager.GetMongoDBCredentials(ctx, secretName)
	if err != nil {
		return nil, fmt.Errorf("failed to get MongoDB credentials: %w", err)
	}

	return &Config{
		MongoURI:     creds.URI,
		DatabaseName: creds.Database,
		Stage:        stage,
	}, nil
}

// loadFromEnv carga la configuración desde variables de entorno
func loadFromEnv(stage string) *Config {
	return &Config{
		MongoURI:     getEnv("MONGODB_URI", ""),
		DatabaseName: getEnv("DATABASE_NAME", "lottery"),
		Stage:        stage,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
