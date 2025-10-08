package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVaultProfilesManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	os.Setenv("HOME", tempDir)

	vpm := NewVaultProfilesManager()

	// Test creating default profiles
	err := vpm.createDefaultProfiles()
	require.NoError(t, err)

	// Test loading profiles
	err = vpm.LoadProfiles()
	require.NoError(t, err)

	// Test getting default profile
	profile, err := vpm.GetProfile("default")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8200", profile.Address)
	assert.Equal(t, "token", profile.AuthMethod)

	// Test listing profiles
	profiles := vpm.ListProfiles()
	assert.Contains(t, profiles, "default")

	// Test setting a new profile
	newProfile := &VaultProfile{
		Address:    "https://test.vault.com",
		AuthMethod: "token",
		Token:      "test-token",
		Namespace:  "test",
	}

	vpm.SetProfile("test", newProfile)

	// Test getting the new profile
	retrievedProfile, err := vpm.GetProfile("test")
	require.NoError(t, err)
	assert.Equal(t, "https://test.vault.com", retrievedProfile.Address)
	assert.Equal(t, "test-token", retrievedProfile.Token)

	// Test saving profiles
	err = vpm.SaveProfiles()
	require.NoError(t, err)

	// Verify file was created
	configPath := vpm.GetConfigPath()
	assert.FileExists(t, configPath)
}

func TestValidateProfile(t *testing.T) {
	vpm := NewVaultProfilesManager()

	// Test valid profile
	validProfile := &VaultProfile{
		Address:    "https://vault.example.com",
		AuthMethod: "token",
		Token:      "valid-token",
	}

	err := vpm.ValidateProfile(validProfile)
	assert.NoError(t, err)

	// Test invalid profile - missing address
	invalidProfile := &VaultProfile{
		AuthMethod: "token",
		Token:      "valid-token",
	}

	err = vpm.ValidateProfile(invalidProfile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vault address is required")

	// Test invalid profile - missing auth method
	invalidProfile2 := &VaultProfile{
		Address: "https://vault.example.com",
		Token:   "valid-token",
	}

	err = vpm.ValidateProfile(invalidProfile2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth method is required")

	// Test invalid profile - token auth without token
	invalidProfile3 := &VaultProfile{
		Address:    "https://vault.example.com",
		AuthMethod: "token",
	}

	err = vpm.ValidateProfile(invalidProfile3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is required")

	// Test invalid profile - unsupported auth method
	invalidProfile4 := &VaultProfile{
		Address:    "https://vault.example.com",
		AuthMethod: "unsupported",
	}

	err = vpm.ValidateProfile(invalidProfile4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported auth method")
}

func TestDeleteProfile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the config path for testing
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	os.Setenv("HOME", tempDir)

	vpm := NewVaultProfilesManager()

	// Create default profiles
	err := vpm.createDefaultProfiles()
	require.NoError(t, err)

	// Add a test profile
	testProfile := &VaultProfile{
		Address:    "https://test.vault.com",
		AuthMethod: "token",
		Token:      "test-token",
	}
	vpm.SetProfile("test", testProfile)

	// Verify profile exists
	_, err = vpm.GetProfile("test")
	require.NoError(t, err)

	// Delete the profile
	err = vpm.DeleteProfile("test")
	require.NoError(t, err)

	// Verify profile is gone
	_, err = vpm.GetProfile("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test deleting non-existent profile
	err = vpm.DeleteProfile("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
