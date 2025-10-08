package ui

import (
	"testing"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		App: config.AppConfig{
			DefaultVault: "test",
		},
		Vault: config.VaultConfig{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
		},
	}

	// Create test vault manager
	vaultMgr, err := vault.NewManager(&cfg.Vault)
	// This will fail because there's no vault server, but we can still test the UI creation
	if err != nil {
		t.Skip("Skipping test - no vault server available")
	}

	logger := logrus.New()

	// Create UI app
	uiApp := NewApp(cfg, vaultMgr, logger)

	assert.NotNil(t, uiApp)
	assert.NotNil(t, uiApp.GetUIApp())
	assert.NotNil(t, uiApp.GetLayout())
}

func TestAppStructure(t *testing.T) {
	// Test that the app structure is correct
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{} // This would normally be properly initialized
	logger := logrus.New()

	uiApp := NewApp(cfg, vaultMgr, logger)

	// Test that we can get the underlying components
	assert.NotNil(t, uiApp.GetUIApp())
	assert.NotNil(t, uiApp.GetLayout())
}
