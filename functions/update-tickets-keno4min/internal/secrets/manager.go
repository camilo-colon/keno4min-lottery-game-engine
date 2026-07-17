package secrets

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Manager struct {
	client *secretsmanager.Client
}

type MongoDBCredentials struct {
	URI      string `json:"uri"`
	Database string `json:"database"`
}

// NewManager crea un nuevo manager de secrets
func NewManager(ctx context.Context) (*Manager, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Manager{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}

// GetMongoDBCredentials obtiene las credenciales de MongoDB desde Secrets Manager
func (m *Manager) GetMongoDBCredentials(ctx context.Context, secretArn string) (*MongoDBCredentials, error) {
	result, err := m.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// SecretString viene vacío si el secret se guardó como binario: el valor
	// estaría en SecretBinary y desreferenciarlo haría panic.
	if result.SecretString == nil {
		return nil, fmt.Errorf("secret %s has no string value (stored as binary?)", secretArn)
	}

	var creds MongoDBCredentials
	if err := json.Unmarshal([]byte(*result.SecretString), &creds); err != nil {
		return nil, fmt.Errorf("failed to parse secret: %w", err)
	}

	return &creds, nil
}
