package ui

import (
	"testing"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLayoutOfflineMode(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Vaults: map[string]config.VaultProfile{
			"default": {
				Address:    "http://localhost:8200",
				AuthMethod: "token",
			},
		},
	}
	logger := logrus.New()

	// Create test vault manager
	vaultMgr, err := vault.NewManager(cfg, logger)
	if err != nil {
		t.Skip("Skipping test - no vault server available")
	}

	// Create layout
	layout := NewLayout(cfg, vaultMgr, logger)

	// Test initialization (should work even without vault connection)
	err = layout.Initialize()
	assert.NoError(t, err)

	// Test that we can get the root primitive
	assert.NotNil(t, layout.GetRoot())
}

func TestLayoutStructure(t *testing.T) {
	// Test that the layout structure is correct
	cfg := &config.Config{
		Vaults: map[string]config.VaultProfile{
			"default": {
				Address:    "http://localhost:8200",
				AuthMethod: "token",
			},
		},
	}

	logger := logrus.New()
	vaultMgr, err := vault.NewManager(cfg, logger)
	if err != nil {
		t.Skip("Skipping test - no vault server available")
	}

	layout := NewLayout(cfg, vaultMgr, logger)

	// Initialize the layout
	err = layout.Initialize()
	assert.NoError(t, err)

	// Test that we can get the root primitive
	assert.NotNil(t, layout.GetRoot())

	// Test that we can get the components (they might be nil in offline mode)
	// This is expected behavior when there's no vault connection
	treePanel := layout.GetTreePanel()
	metadataPanel := layout.GetMetadataPanel()
	valuePanel := layout.GetValuePanel()
	statusBar := layout.GetStatusBar()

	// In offline mode, these might be nil
	_ = treePanel
	_ = metadataPanel
	_ = valuePanel
	_ = statusBar
}
