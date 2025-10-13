package handlers

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/rvolykh/vui/internal/vault"
)

// ClipboardWriter is an interface for writing to clipboard (for testing)
type ClipboardWriter interface {
	WriteAll(text string) error
}

// RealClipboard uses the actual clipboard
type RealClipboard struct{}

// WriteAll writes to the real clipboard
func (rc *RealClipboard) WriteAll(text string) error {
	return clipboard.WriteAll(text)
}

// ClipboardHandler handles clipboard operations
type ClipboardHandler struct {
	vaultMgr *vault.Manager
	writer   ClipboardWriter
}

// NewClipboardHandler creates a new clipboard handler with the real clipboard
func NewClipboardHandler(vaultMgr *vault.Manager) *ClipboardHandler {
	return &ClipboardHandler{
		vaultMgr: vaultMgr,
		writer:   &RealClipboard{},
	}
}

// NewClipboardHandlerWithWriter creates a clipboard handler with a custom writer (for testing)
func NewClipboardHandlerWithWriter(vaultMgr *vault.Manager, writer ClipboardWriter) *ClipboardHandler {
	return &ClipboardHandler{
		vaultMgr: vaultMgr,
		writer:   writer,
	}
}

// CopyKeyValue copies a specific key's value from a secret to clipboard
func (ch *ClipboardHandler) CopyKeyValue(secret *vault.SecretNode, key string) error {
	if secret == nil {
		return fmt.Errorf("secret is nil")
	}

	if key == "" {
		return fmt.Errorf("key is required")
	}

	// Get the full secret data if not already loaded
	if secret.Data == nil {
		secretsManager, err := ch.vaultMgr.GetSecretsManager()
		if err != nil {
			return fmt.Errorf("failed to get secrets manager: %w", err)
		}

		fullSecret, err := secretsManager.GetSecret(secret.Path)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}
		secret = fullSecret
	}

	value, ok := secret.Data[key]
	if !ok {
		return fmt.Errorf("key '%s' not found in secret", key)
	}

	valueStr := fmt.Sprintf("%v", value)
	if err := ch.writer.WriteAll(valueStr); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// CopySecretValues copies all values from a secret to clipboard
// If there's only one value, copies just that value
// If there are multiple values, copies them one per line
func (ch *ClipboardHandler) CopySecretValues(secret *vault.SecretNode) error {
	if secret == nil {
		return fmt.Errorf("secret is nil")
	}

	// Get the full secret data if not already loaded
	if secret.Data == nil {
		secretsManager, err := ch.vaultMgr.GetSecretsManager()
		if err != nil {
			return fmt.Errorf("failed to get secrets manager: %w", err)
		}

		fullSecret, err := secretsManager.GetSecret(secret.Path)
		if err != nil {
			return fmt.Errorf("failed to get secret: %w", err)
		}
		secret = fullSecret
	}

	if len(secret.Data) == 0 {
		return fmt.Errorf("no data in secret")
	}

	// If there's only one key, copy just the value
	if len(secret.Data) == 1 {
		for _, value := range secret.Data {
			valueStr := fmt.Sprintf("%v", value)
			if err := ch.writer.WriteAll(valueStr); err != nil {
				return fmt.Errorf("failed to copy to clipboard: %w", err)
			}
			return nil
		}
	}

	// Multiple values - copy all values (one per line)
	var values []string
	for _, value := range secret.Data {
		values = append(values, fmt.Sprintf("%v", value))
	}
	valueStr := strings.Join(values, "\n")

	if err := ch.writer.WriteAll(valueStr); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}
