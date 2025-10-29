package panels

import (
	"testing"
	"time"

	"github.com/rvolykh/vui/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretsValue(t *testing.T) {
	fixtures := WithFixtures(t)

	panel := NewSecretsValue(fixtures.cfg, fixtures.interactor, fixtures.logger)

	assert.NotNil(t, panel)
	assert.True(t, panel.isMasked) // Should start masked
}

func TestSecretsValue_Initialize(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsValue(fixtures.cfg, fixtures.interactor, fixtures.logger)

	err := panel.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, panel.GetPrimitive())
}

func TestSecretsValue_ShowSecret(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsValue(fixtures.cfg, fixtures.interactor, fixtures.logger)
	panel.Initialize()
	secret := &models.SecretNode{
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

func TestSecretsValue_ShowKey(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsValue(fixtures.cfg, fixtures.interactor, fixtures.logger)
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
	assert.Equal(t, "password", panel.currentKey)
	assert.True(t, panel.isMasked) // Should reset to masked
}

func TestSecretsValue_ToggleMasking(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsValue(fixtures.cfg, fixtures.interactor, fixtures.logger)
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
	assert.True(t, panel.isMasked)

	panel.ToggleMasking()
	assert.False(t, panel.isMasked)

	panel.ToggleMasking()
	assert.True(t, panel.isMasked)
}

func TestSecretsValue_ShowDirectory(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsValue(fixtures.cfg, fixtures.interactor, fixtures.logger)
	panel.Initialize()

	panel.ShowDirectory("secrets/test")

	assert.Nil(t, panel.currentSecret)
	assert.Equal(t, "", panel.currentKey)
}

func TestSecretsValue_FormatValue(t *testing.T) {
	fixtures := WithFixtures(t)
	panel := NewSecretsValue(fixtures.cfg, fixtures.interactor, fixtures.logger)

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
