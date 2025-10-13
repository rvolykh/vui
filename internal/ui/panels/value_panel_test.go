package panels

import (
	"testing"
	"time"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewValuePanel(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewValuePanel(cfg, vaultMgr, logger)

	assert.NotNil(t, panel)
	assert.True(t, panel.isMasked) // Should start masked
}

func TestValuePanel_Initialize(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewValuePanel(cfg, vaultMgr, logger)
	err := panel.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, panel.GetPrimitive())
}

func TestValuePanel_ShowSecret(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewValuePanel(cfg, vaultMgr, logger)
	panel.Initialize()

	secret := &vault.SecretNode{
		Name:     "test-secret",
		Path:     "secrets/test",
		IsSecret: true,
		Data: map[string]interface{}{
			"password": "secret123",
			"username": "admin",
		},
	}

	panel.ShowSecret(secret)

	assert.Equal(t, secret, panel.currentSecret)
	assert.Equal(t, "", panel.currentKey)
	assert.True(t, panel.isMasked) // Should reset to masked
}

func TestValuePanel_ShowKey(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewValuePanel(cfg, vaultMgr, logger)
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
	assert.Equal(t, "password", panel.currentKey)
	assert.True(t, panel.isMasked) // Should reset to masked
}

func TestValuePanel_ToggleMasking(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewValuePanel(cfg, vaultMgr, logger)
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
	assert.True(t, panel.isMasked)

	panel.ToggleMasking()
	assert.False(t, panel.isMasked)

	panel.ToggleMasking()
	assert.True(t, panel.isMasked)
}

func TestValuePanel_ShowDirectory(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewValuePanel(cfg, vaultMgr, logger)
	panel.Initialize()

	panel.ShowDirectory("secrets/test")

	assert.Nil(t, panel.currentSecret)
	assert.Equal(t, "", panel.currentKey)
}

func TestValuePanel_FormatValue(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	vaultMgr, _ := vault.NewManager(cfg, logger)

	panel := NewValuePanel(cfg, vaultMgr, logger)

	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "string value",
			value:    "test string",
			expected: "test string",
		},
		{
			name:     "byte array",
			value:    []byte("byte array"),
			expected: "byte array",
		},
		{
			name:     "integer",
			value:    123,
			expected: "123",
		},
		{
			name:     "boolean",
			value:    true,
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := panel.formatValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
