package engines

import (
	"context"

	"github.com/rvolykh/vui/internal/models"
)

type SecretEngine interface {
	SecretClient

	Authenticate() error
	GetAddress() string
	GetStatus(ctx context.Context) (models.ConnectionStatus, error)
}

type SecretClient interface {
	ListSecrets(path string) ([]*models.SecretNode, error)
	GetSecret(path string) (*models.SecretNode, error)
	CreateSecret(path string, data map[string]any) error
	UpdateSecret(path string, data map[string]any) error
	DeleteSecret(path string) error
}
