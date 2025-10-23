package vault

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
)

// Client wraps the Vault API client with additional functionality
type Client struct {
	apiClient *api.Client
	profile   *config.VaultProfile
	logger    *logrus.Logger
}

// TestConnection tests the connection to the vault
func (c *Client) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to get the vault status
	status, err := c.apiClient.Sys().SealStatusWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get vault status: %w", err)
	}

	c.logger.Infof("Connected to vault at %s (sealed: %v)", c.apiClient.Address(), status.Sealed)
	return nil
}

// GetSecret retrieves a secret from the vault
func (c *Client) GetSecret(path string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secret, err := c.apiClient.Logical().ReadWithContext(ctx, "secret/data/"+path)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret at path '%s': %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path '%s'", path)
	}

	// For KV v2, the actual data is nested under "data"
	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		return data, nil
	}

	return secret.Data, nil
}

// ListSecrets lists secrets at a given path
func (c *Client) ListSecrets(path string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secret, err := c.apiClient.Logical().ListWithContext(ctx, "secret/metadata/"+path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets at path '%s': %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	// Extract keys from the response
	if keys, ok := secret.Data["keys"].([]interface{}); ok {
		result := make([]string, len(keys))
		for i, key := range keys {
			if keyStr, ok := key.(string); ok {
				result[i] = keyStr
			}
		}
		return result, nil
	}

	return []string{}, nil
}

// CreateSecret creates a new secret in the vault
func (c *Client) CreateSecret(path string, data map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// For KV v2, we need to wrap the data
	secretData := map[string]interface{}{
		"data": data,
	}

	_, err := c.apiClient.Logical().WriteWithContext(ctx, "secret/data/"+path, secretData)
	if err != nil {
		return fmt.Errorf("failed to create secret at path '%s': %w", path, err)
	}

	c.logger.Infof("Created secret at path: %s", path)
	return nil
}

// UpdateSecret updates an existing secret in the vault
func (c *Client) UpdateSecret(path string, data map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// For KV v2, we need to wrap the data
	secretData := map[string]interface{}{
		"data": data,
	}

	_, err := c.apiClient.Logical().WriteWithContext(ctx, "secret/data/"+path, secretData)
	if err != nil {
		return fmt.Errorf("failed to update secret at path '%s': %w", path, err)
	}

	c.logger.Infof("Updated secret at path: %s", path)
	return nil
}

// DeleteSecret deletes a secret from the vault
func (c *Client) DeleteSecret(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.apiClient.Logical().DeleteWithContext(ctx, "secret/metadata/"+path)
	if err != nil {
		return fmt.Errorf("failed to delete secret at path '%s': %w", path, err)
	}

	c.logger.Infof("Deleted secret at path: %s", path)
	return nil
}

// GetVaultInfo returns information about the vault
func (c *Client) GetVaultInfo() (*VaultInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := c.apiClient.Sys().SealStatusWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault status: %w", err)
	}

	health, err := c.apiClient.Sys().HealthWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault health: %w", err)
	}

	return &VaultInfo{
		Address:     c.apiClient.Address(),
		Sealed:      status.Sealed,
		Version:     status.Version,
		ClusterID:   status.ClusterID,
		ClusterName: status.ClusterName,
		Initialized: health.Initialized,
		Standby:     health.Standby,
	}, nil
}

// VaultInfo contains information about the vault
type VaultInfo struct {
	Address     string `json:"address"`
	Sealed      bool   `json:"sealed"`
	Version     string `json:"version"`
	ClusterID   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
	Initialized bool   `json:"initialized"`
	Standby     bool   `json:"standby"`
}
