package panels

import (
	"testing"
	"time"

	"github.com/rvolykh/vui/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretsMetadata(t *testing.T) {
	fixtures := WithFixtures(t)

	panel := NewSecretsMetadata(fixtures.cfg, fixtures.interactor, fixtures.logger)

	assert.NotNil(t, panel)
}

func TestSecretsMetadata_Initialize(t *testing.T) {
	fixtures := WithFixtures(t)

	panel := NewSecretsMetadata(fixtures.cfg, fixtures.interactor, fixtures.logger)
	err := panel.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, panel.GetPrimitive())
}

func TestSecretsMetadata_ShowSecret(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsMetadata(fixtures.cfg, fixtures.interactor, fixtures.logger)
	panel.Initialize()

	secret := &models.SecretNode{
		Name:     "test-secret",
		Path:     "secrets/test",
		IsSecret: true,
		Data: map[string]interface{}{
			"password": "secret123",
			"username": "admin",
		},
		Metadata: &models.SecretMetadata{
			Version:     2,
			CreatedTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Destroyed:   false,
		},
	}

	panel.ShowSecret(secret)

	assert.Equal(t, secret, panel.currentSecret)
}

func TestSecretsMetadata_ShowDirectory(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsMetadata(fixtures.cfg, fixtures.interactor, fixtures.logger)
	panel.Initialize()

	panel.ShowDirectory("secrets/test", 5)

	assert.Nil(t, panel.currentSecret)
}

func TestSecretsMetadata_ShowKey(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsMetadata(fixtures.cfg, fixtures.interactor, fixtures.logger)
	panel.Initialize()
	secret := &models.SecretNode{
		Name:     "test-secret",
		Path:     "secrets/test",
		IsSecret: true,
		Data: map[string]interface{}{
			"password": "secret123",
		},
	}

	panel.ShowKey(secret, "password")

	assert.Equal(t, secret, panel.currentSecret)
}

func TestSecretsMetadata_Refresh(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsMetadata(fixtures.cfg, fixtures.interactor, fixtures.logger)
	panel.Initialize()
	secret := &models.SecretNode{
		Name:     "test-secret",
		Path:     "secrets/test",
		IsSecret: true,
		Data: map[string]interface{}{
			"password": "secret123",
		},
		Metadata: &models.SecretMetadata{
			Version:     1,
			CreatedTime: time.Now(),
		},
	}

	panel.ShowSecret(secret)

	// Should not panic
	panel.Refresh()

	assert.Equal(t, secret, panel.currentSecret)
}
