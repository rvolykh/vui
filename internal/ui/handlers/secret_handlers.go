package handlers

import (
	"fmt"

	"github.com/rvolykh/vui/internal/backend"
	"github.com/rvolykh/vui/internal/models"
)

// SecretHandler handles secret operations
type SecretHandler struct {
	interactor backend.Interactor
}

// NewSecretHandler creates a new secret handler
func NewSecretHandler(interactor backend.Interactor) *SecretHandler {
	return &SecretHandler{
		interactor: interactor,
	}
}

// CreateSecret creates a new secret with the given key-value pairs
func (sh *SecretHandler) CreateSecret(path string, data map[string]interface{}) error {
	if path == "" {
		return fmt.Errorf("secret path is required")
	}

	if len(data) == 0 {
		return fmt.Errorf("at least one key-value pair is required")
	}

	secretsInteractor, err := sh.interactor.Secrets()
	if err != nil {
		return fmt.Errorf("failed to get secrets manager: %w", err)
	}

	if err := secretsInteractor.CreateSecret(path, data); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// UpdateSecret updates an existing secret with the given key-value pairs
func (sh *SecretHandler) UpdateSecret(path string, data map[string]interface{}) error {
	if path == "" {
		return fmt.Errorf("secret path is required")
	}

	if len(data) == 0 {
		return fmt.Errorf("at least one key-value pair is required")
	}

	secretsInteractor, err := sh.interactor.Secrets()
	if err != nil {
		return fmt.Errorf("failed to get secrets manager: %w", err)
	}

	if err := secretsInteractor.UpdateSecret(path, data); err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// DeleteSecret deletes a secret at the given path
func (sh *SecretHandler) DeleteSecret(path string) error {
	if path == "" {
		return fmt.Errorf("secret path is required")
	}

	secretsInteractor, err := sh.interactor.Secrets()
	if err != nil {
		return fmt.Errorf("failed to get secrets manager: %w", err)
	}

	if err := secretsInteractor.DeleteSecret(path); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// GetSecret retrieves a secret at the given path
func (sh *SecretHandler) GetSecret(path string) (*models.SecretNode, error) {
	if path == "" {
		return nil, fmt.Errorf("secret path is required")
	}

	secretsInteractor, err := sh.interactor.Secrets()
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets manager: %w", err)
	}

	secret, err := secretsInteractor.GetSecret(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret, nil
}

// ListSecrets lists all secrets at the given path
func (sh *SecretHandler) ListSecrets(path string) ([]*models.SecretNode, error) {
	secretsInteractor, err := sh.interactor.Secrets()
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets manager: %w", err)
	}

	secrets, err := secretsInteractor.ListSecrets(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secrets, nil
}

// CreateSecretWithCallbacks creates a secret with success/error callbacks
func (sh *SecretHandler) CreateSecretWithCallbacks(path string, data map[string]interface{}, onSuccess func(), onError func(error)) {
	err := sh.CreateSecret(path, data)
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return
	}

	if onSuccess != nil {
		onSuccess()
	}
}

// UpdateSecretWithCallbacks updates a secret with success/error callbacks
func (sh *SecretHandler) UpdateSecretWithCallbacks(path string, data map[string]interface{}, onSuccess func(), onError func(error)) {
	err := sh.UpdateSecret(path, data)
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return
	}

	if onSuccess != nil {
		onSuccess()
	}
}

// DeleteSecretWithCallbacks deletes a secret with success/error callbacks
func (sh *SecretHandler) DeleteSecretWithCallbacks(path string, onSuccess func(), onError func(error)) {
	err := sh.DeleteSecret(path)
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return
	}

	if onSuccess != nil {
		onSuccess()
	}
}
