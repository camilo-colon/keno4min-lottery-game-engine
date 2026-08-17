package secrets

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Manager struct {
	client *secretsmanager.Client
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

// GetSecret obtiene el secret como string.
func (m *Manager) GetSecret(ctx context.Context, secretArn string) (string, error) {
	result, err := m.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret %s has no string value", secretArn)
	}

	return *result.SecretString, nil
}
