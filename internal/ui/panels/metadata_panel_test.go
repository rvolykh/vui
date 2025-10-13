package panels

import (
	"testing"
	"time"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewMetadataPanel(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewMetadataPanel(cfg, vaultMgr, logger)

	assert.NotNil(t, panel)
}

func TestMetadataPanel_Initialize(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewMetadataPanel(cfg, vaultMgr, logger)
	err := panel.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, panel.GetPrimitive())
}

func TestMetadataPanel_ShowSecret(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewMetadataPanel(cfg, vaultMgr, logger)
	panel.Initialize()

	secret := &vault.SecretNode{
		Name:     "test-secret",
		Path:     "secrets/test",
		IsSecret: true,
		Data: map[string]interface{}{
			"password": "secret123",
			"username": "admin",
		},
		Metadata: &vault.SecretMetadata{
			Version:     2,
			CreatedTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Destroyed:   false,
		},
	}

	panel.ShowSecret(secret)

	assert.Equal(t, secret, panel.currentSecret)
}

func TestMetadataPanel_ShowDirectory(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewMetadataPanel(cfg, vaultMgr, logger)
	panel.Initialize()

	panel.ShowDirectory("secrets/test", 5)

	assert.Nil(t, panel.currentSecret)
}

func TestMetadataPanel_ShowKey(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewMetadataPanel(cfg, vaultMgr, logger)
	panel.Initialize()

	secret := &vault.SecretNode{
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

func TestMetadataPanel_Refresh(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewMetadataPanel(cfg, vaultMgr, logger)
	panel.Initialize()

	secret := &vault.SecretNode{
		Name:     "test-secret",
		Path:     "secrets/test",
		IsSecret: true,
		Data: map[string]interface{}{
			"password": "secret123",
		},
		Metadata: &vault.SecretMetadata{
			Version:     1,
			CreatedTime: time.Now(),
		},
	}

	panel.ShowSecret(secret)

	// Should not panic
	panel.Refresh()

	assert.Equal(t, secret, panel.currentSecret)
}
