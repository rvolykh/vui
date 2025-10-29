package panels

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecretsStatus(t *testing.T) {
	fixtures := WithFixtures(t)

	statusBar := NewSecretsStatus(fixtures.cfg, fixtures.interactor, fixtures.logger)

	assert.NotNil(t, statusBar)
}

func TestSecretsStatus_Initialize(t *testing.T) {
	fixtures := WithFixtures(t)

	statusBar := NewSecretsStatus(fixtures.cfg, fixtures.interactor, fixtures.logger)
	err := statusBar.Initialize()

	assert.NoError(t, err)
	assert.NotNil(t, statusBar.GetPrimitive())
}

func TestSecretsStatus_UpdateSelection(t *testing.T) {
	fixtures := WithFixtures(t)

	statusBar := NewSecretsStatus(fixtures.cfg, fixtures.interactor, fixtures.logger)
	statusBar.Initialize()

	tests := []struct {
		name     string
		path     string
		isSecret bool
	}{
		{
			name:     "secret selection",
			path:     "secrets/test/password",
			isSecret: true,
		},
		{
			name:     "directory selection",
			path:     "secrets/test",
			isSecret: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			statusBar.UpdateSelection(tt.path, tt.isSecret)
		})
	}
}
