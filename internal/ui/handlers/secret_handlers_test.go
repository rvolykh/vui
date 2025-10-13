package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecretHandler(t *testing.T) {
	handler := NewSecretHandler(nil)
	assert.NotNil(t, handler)
}

func TestSecretHandler_CreateSecret_ValidationErrors(t *testing.T) {
	handler := NewSecretHandler(nil)

	tests := []struct {
		name        string
		path        string
		data        map[string]interface{}
		expectedErr string
	}{
		{
			name:        "empty path",
			path:        "",
			data:        map[string]interface{}{"key": "value"},
			expectedErr: "secret path is required",
		},
		{
			name:        "empty data",
			path:        "test/secret",
			data:        map[string]interface{}{},
			expectedErr: "at least one key-value pair is required",
		},
		{
			name:        "nil data",
			path:        "test/secret",
			data:        nil,
			expectedErr: "at least one key-value pair is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.CreateSecret(tt.path, tt.data)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestSecretHandler_UpdateSecret_ValidationErrors(t *testing.T) {
	handler := NewSecretHandler(nil)

	tests := []struct {
		name        string
		path        string
		data        map[string]interface{}
		expectedErr string
	}{
		{
			name:        "empty path",
			path:        "",
			data:        map[string]interface{}{"key": "value"},
			expectedErr: "secret path is required",
		},
		{
			name:        "empty data",
			path:        "test/secret",
			data:        map[string]interface{}{},
			expectedErr: "at least one key-value pair is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.UpdateSecret(tt.path, tt.data)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestSecretHandler_DeleteSecret_ValidationErrors(t *testing.T) {
	handler := NewSecretHandler(nil)

	err := handler.DeleteSecret("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret path is required")
}

func TestSecretHandler_GetSecret_ValidationErrors(t *testing.T) {
	handler := NewSecretHandler(nil)

	_, err := handler.GetSecret("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret path is required")
}
