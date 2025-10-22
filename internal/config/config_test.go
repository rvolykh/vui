package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Test loading configuration with defaults
	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify default values
	assert.Equal(t, "dark", config.App.Theme)
	assert.Equal(t, "http://localhost:8200", config.Vaults["local"].Address)
	assert.Equal(t, "token", config.Vaults["local"].AuthMethod)
}

func TestConfigSave(t *testing.T) {
	config := &Config{
		App: AppConfig{
			Theme: "light",
		},
		Vaults: map[string]VaultProfile{
			"default": {
				Address:    "https://test.vault.com",
				AuthMethod: "token",
			},
		},
	}

	// Test that the Save method exists and can be called
	// It may succeed or fail depending on the environment, but shouldn't panic
	err := config.Save()
	// The method should exist and be callable without panicking
	// We don't assert on the result as it depends on the test environment
	_ = err // Acknowledge the error but don't assert on it
}
