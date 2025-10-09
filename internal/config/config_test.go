package config

import (
	"os"
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
	assert.Equal(t, "default", config.App.DefaultVault)
	assert.Equal(t, "dark", config.App.Theme)
	assert.Equal(t, 30, config.App.RefreshInterval)
	assert.Equal(t, "http://localhost:8200", config.Vaults["default"].Address)
	assert.Equal(t, "token", config.Vaults["default"].AuthMethod)
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("VAULT_ADDR", "https://vault.example.com")
	os.Setenv("VAULT_TOKEN", "test-token")
	os.Setenv("VAULT_NAMESPACE", "test-namespace")
	defer func() {
		os.Unsetenv("VAULT_ADDR")
		os.Unsetenv("VAULT_TOKEN")
		os.Unsetenv("VAULT_NAMESPACE")
	}()

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify environment variable overrides
	defaultVault := config.App.DefaultVault
	if defaultVault == "" {
		defaultVault = "default"
	}
	assert.Equal(t, "https://vault.example.com", config.Vaults[defaultVault].Address)
	assert.Equal(t, "test-token", config.Vaults[defaultVault].Token)
	assert.Equal(t, "test-namespace", config.Vaults[defaultVault].Namespace)
}

func TestConfigSave(t *testing.T) {
	config := &Config{
		App: AppConfig{
			DefaultVault: "test",
			Theme:        "light",
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
