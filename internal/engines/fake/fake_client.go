package fake

import (
	"context"

	"github.com/rvolykh/vui/internal/models"
)

type FakeClient struct {
	RespGetStatus   models.ConnectionStatus
	RespListSecrets []*models.SecretNode
	RespGetSecret   *models.SecretNode
	RespGetAddress  string
	RespErr         error
}

func NewFakeClient() *FakeClient {
	return &FakeClient{}
}

func (c *FakeClient) ListSecrets(path string) ([]*models.SecretNode, error) {
	return c.RespListSecrets, c.RespErr
}

func (c *FakeClient) GetSecret(path string) (*models.SecretNode, error) {
	return c.RespGetSecret, c.RespErr
}

func (c *FakeClient) CreateSecret(path string, data map[string]any) error {
	return c.RespErr
}

func (c *FakeClient) UpdateSecret(path string, data map[string]any) error {
	return c.RespErr
}

func (c *FakeClient) DeleteSecret(path string) error {
	return c.RespErr
}

func (c *FakeClient) GetStatus(ctx context.Context) (models.ConnectionStatus, error) {
	return c.RespGetStatus, c.RespErr
}

func (c *FakeClient) GetAddress() string {
	return c.RespGetAddress
}

func (c *FakeClient) Authenticate() error {
	return c.RespErr
}
