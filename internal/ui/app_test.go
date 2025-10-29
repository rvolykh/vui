package ui

import (
	"testing"

	"github.com/rvolykh/vui/internal/backend"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		App: config.AppConfig{},
		Vaults: map[string]config.VaultProfile{
			"test": {
				Address:    "http://localhost:8200",
				AuthMethod: "token",
			},
		},
	}

	logger := logrus.New()
	// Create test interactor
	interactor, err := backend.NewInteractor(logger, cfg)
	// This will fail because there's no vault server, but we can still test the UI creation
	if err != nil {
		t.Skip("Skipping test - no vault server available")
	}

	// Create UI app
	uiApp := NewApp(cfg, interactor, logger)

	assert.NotNil(t, uiApp)
	assert.NotNil(t, uiApp.GetUIApp())
	assert.NotNil(t, uiApp.GetLayout())
}

func TestAppStructure(t *testing.T) {
	// Test that the app structure is correct
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, _ := backend.NewInteractor(logger, cfg)

	uiApp := NewApp(cfg, interactor, logger)

	// Test that we can get the underlying components
	assert.NotNil(t, uiApp.GetUIApp())
	assert.NotNil(t, uiApp.GetLayout())
}
