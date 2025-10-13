package handlers

import (
	"testing"

	"github.com/rvolykh/vui/internal/vault"
	"github.com/stretchr/testify/assert"
)

// MockClipboard is a mock clipboard for testing
type MockClipboard struct {
	content string
	err     error
}

func (mc *MockClipboard) WriteAll(text string) error {
	if mc.err != nil {
		return mc.err
	}
	mc.content = text
	return nil
}

func TestNewClipboardHandler(t *testing.T) {
	handler := NewClipboardHandler(nil)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.writer)
}

func TestNewClipboardHandlerWithWriter(t *testing.T) {
	mockClip := &MockClipboard{}
	handler := NewClipboardHandlerWithWriter(nil, mockClip)
	assert.NotNil(t, handler)
	assert.Equal(t, mockClip, handler.writer)
}

func TestClipboardHandler_CopyKeyValue(t *testing.T) {
	mockClip := &MockClipboard{}
	handler := NewClipboardHandlerWithWriter(nil, mockClip)

	tests := []struct {
		name        string
		secret      *vault.SecretNode
		key         string
		expectError bool
		expectedVal string
	}{
		{
			name:        "nil secret",
			secret:      nil,
			key:         "password",
			expectError: true,
		},
		{
			name: "empty key",
			secret: &vault.SecretNode{
				Name: "test",
				Path: "secrets/test",
				Data: map[string]interface{}{
					"password": "secret123",
				},
			},
			key:         "",
			expectError: true,
		},
		{
			name: "key not found",
			secret: &vault.SecretNode{
				Name: "test",
				Path: "secrets/test",
				Data: map[string]interface{}{
					"password": "secret123",
				},
			},
			key:         "nonexistent",
			expectError: true,
		},
		{
			name: "successful copy",
			secret: &vault.SecretNode{
				Name: "test",
				Path: "secrets/test",
				Data: map[string]interface{}{
					"password": "secret123",
				},
			},
			key:         "password",
			expectError: false,
			expectedVal: "secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClip.content = ""
			err := handler.CopyKeyValue(tt.secret, tt.key)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVal, mockClip.content)
			}
		})
	}
}

func TestClipboardHandler_CopySecretValues(t *testing.T) {
	mockClip := &MockClipboard{}
	handler := NewClipboardHandlerWithWriter(nil, mockClip)

	tests := []struct {
		name        string
		secret      *vault.SecretNode
		expectError bool
		validate    func(t *testing.T, content string)
	}{
		{
			name:        "nil secret",
			secret:      nil,
			expectError: true,
		},
		{
			name: "empty data",
			secret: &vault.SecretNode{
				Name: "test",
				Path: "secrets/test",
				Data: map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "single value",
			secret: &vault.SecretNode{
				Name: "test",
				Path: "secrets/test",
				Data: map[string]interface{}{
					"password": "secret123",
				},
			},
			expectError: false,
			validate: func(t *testing.T, content string) {
				assert.Equal(t, "secret123", content)
			},
		},
		{
			name: "multiple values",
			secret: &vault.SecretNode{
				Name: "test",
				Path: "secrets/test",
				Data: map[string]interface{}{
					"username": "admin",
					"password": "secret123",
				},
			},
			expectError: false,
			validate: func(t *testing.T, content string) {
				// Content should contain both values separated by newline
				assert.Contains(t, content, "admin")
				assert.Contains(t, content, "secret123")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClip.content = ""
			err := handler.CopySecretValues(tt.secret)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, mockClip.content)
				}
			}
		})
	}
}
