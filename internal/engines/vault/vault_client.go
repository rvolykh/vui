package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/engines/vault/auth"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
)

type VaultClient struct {
	apiClient *api.Client
	profile   *config.Profile
	logger    *logrus.Logger
}

func NewVaultClient(logger *logrus.Logger, profile *config.Profile) (*VaultClient, error) {
	apiConfig := api.DefaultConfig()
	apiConfig.Address = profile.Address

	if profile.CertPath != "" {
		if err := apiConfig.ConfigureTLS(&api.TLSConfig{
			CACert: profile.CertPath,
		}); err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
	}

	apiClient, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Set namespace if provided
	if profile.Namespace != "" {
		apiClient.SetNamespace(profile.Namespace)
	}

	return &VaultClient{
		apiClient: apiClient,
		profile:   profile,
		logger:    logger,
	}, nil
}

func (c *VaultClient) Authenticate() error {
	authManager := auth.NewAuthManager(c.logger)

	if err := authManager.Authenticate(c.apiClient, c.profile); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err := authManager.VerifyAuthentication(c.apiClient); err != nil {
		return fmt.Errorf("failed to verify authentication: %w", err)
	}

	return nil
}

func (c *VaultClient) GetAddress() string {
	return c.apiClient.Address()
}

func (c *VaultClient) GetStatus(ctx context.Context) (models.ConnectionStatus, error) {
	status, err := c.apiClient.Sys().SealStatusWithContext(ctx)
	if err != nil {
		return models.ConnectionStatus{}, fmt.Errorf("failed to get seal status: %w", err)
	}

	if status.Sealed {
		return models.ConnectionStatus{
			Status:    models.StatusSealed,
			Address:   c.GetAddress(),
			Version:   status.Version,
			ClusterID: status.ClusterID,
		}, nil
	}

	return models.ConnectionStatus{
		Status:    models.StatusConnected,
		Address:   c.GetAddress(),
		Version:   status.Version,
		ClusterID: status.ClusterID,
	}, nil
}

func (c *VaultClient) ListSecrets(path string) ([]*models.SecretNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secret, err := c.apiClient.Logical().ListWithContext(ctx, "secret/metadata/"+strings.Trim(path, "/"))
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets at path '%s': %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return []*models.SecretNode{}, nil
	}

	var nodes []*models.SecretNode
	keys, ok := secret.Data["keys"].([]any)
	if !ok {
		return []*models.SecretNode{}, nil
	}
	for _, key := range keys {
		if keyStr, ok := key.(string); ok {
			isSecret := !strings.HasSuffix(keyStr, "/")
			keyStr = strings.TrimSuffix(keyStr, "/")

			node := &models.SecretNode{
				Name:     keyStr,
				Path:     filepath.Join(path, keyStr),
				IsSecret: isSecret,
			}

			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func (c *VaultClient) GetSecret(path string) (*models.SecretNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	data, err := c.apiClient.Logical().ReadWithContext(ctx, "secret/data/"+path)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret at path '%s': %w", path, err)
	}

	if data == nil || data.Data == nil {
		return nil, fmt.Errorf("secret not found at path '%s'", path)
	}

	metadata, err := c.apiClient.Logical().ReadWithContext(ctx, "secret/metadata/"+path)
	if err != nil {
		c.logger.Warnf("Failed to get metadata for secret '%s': %v", path, err)
	}

	node := &models.SecretNode{
		Name:     filepath.Base(path),
		Path:     path,
		IsSecret: true,
		Metadata: &models.SecretMetadata{},
	}

	// Extract data from KV v2 response
	if dataMap, ok := data.Data["data"].(map[string]any); ok {
		node.Data = dataMap
	} else {
		node.Data = data.Data
	}

	if metadata != nil && metadata.Data != nil {
		if createdTime, ok := metadata.Data["created_time"].(string); ok {
			if t, err := time.Parse(time.RFC3339, createdTime); err == nil {
				node.Metadata.CreatedTime = t
			}
		}

		version, ok := metadata.Data["current_version"].(json.Number)
		if !ok {
			return nil, fmt.Errorf("failed to get version for secret '%s'", path)
		}
		versionInt, err := version.Int64()
		if err != nil {
			return nil, fmt.Errorf("failed to get version for secret '%s': %w", path, err)
		}

		node.Metadata.Version = int(versionInt)
	}

	return node, nil
}

func (c *VaultClient) CreateSecret(path string, data map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Handle empty key marker (defensive check - handler should already transform for Vault)
	if value, hasEmptyKey := data[""]; hasEmptyKey && len(data) == 1 {
		data = map[string]any{
			"value": value,
		}
	}

	// For KV v2, we need to wrap the data
	secretData := map[string]any{
		"data": data,
	}

	_, err := c.apiClient.Logical().WriteWithContext(ctx, "secret/data/"+path, secretData)
	if err != nil {
		return fmt.Errorf("failed to create secret at path '%s': %w", path, err)
	}

	c.logger.Infof("Created secret at path: %s", path)
	return nil
}

func (c *VaultClient) UpdateSecret(path string, data map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Handle empty key marker (defensive check - handler should already transform for Vault)
	if value, hasEmptyKey := data[""]; hasEmptyKey && len(data) == 1 {
		data = map[string]any{
			"value": value,
		}
	}

	// For KV v2, we need to wrap the data
	secretData := map[string]any{
		"data": data,
	}

	_, err := c.apiClient.Logical().WriteWithContext(ctx, "secret/data/"+path, secretData)
	if err != nil {
		return fmt.Errorf("failed to update secret at path '%s': %w", path, err)
	}

	c.logger.Infof("Updated secret at path: %s", path)
	return nil
}

func (c *VaultClient) DeleteSecret(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.apiClient.Logical().DeleteWithContext(ctx, "secret/metadata/"+path)
	if err != nil {
		return fmt.Errorf("failed to delete secret at path '%s': %w", path, err)
	}

	c.logger.Infof("Deleted secret at path: %s", path)
	return nil
}
