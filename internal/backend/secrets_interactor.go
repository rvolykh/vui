package backend

import (
	"path/filepath"

	"github.com/rvolykh/vui/internal/engines"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
)

type SecretsInteractor interface {
	engines.SecretClient
	BuildTree(rootPath string, maxDepth int) (*models.SecretNode, error)
}

type secretsInteractor struct {
	name   string
	logger *logrus.Logger
	client engines.SecretClient
	cache  *secretCache
}

func newSecretsInteractor(logger *logrus.Logger, name string, client engines.SecretClient) SecretsInteractor {
	return &secretsInteractor{
		logger: logger,
		name:   name,
		client: client,
		cache:  newSecretCache(),
	}
}

// ListSecrets retrieves secrets at the given path, using cache if available
func (i *secretsInteractor) ListSecrets(path string) ([]*models.SecretNode, error) {
	// Check cache
	if secrets, found := i.cache.GetListSecrets(path); found {
		return secrets, nil
	}

	// Cache miss, fetch from underlying client
	secrets, err := i.client.ListSecrets(path)
	if err != nil {
		return nil, err
	}

	// Store in cache
	i.cache.SetListSecrets(path, secrets)

	return secrets, nil
}

// GetSecret retrieves a secret at the given path, using cache if available
func (i *secretsInteractor) GetSecret(path string) (*models.SecretNode, error) {
	// Check cache
	if secret, found := i.cache.GetSecret(path); found {
		return secret, nil
	}

	// Cache miss, fetch from underlying client
	secret, err := i.client.GetSecret(path)
	if err != nil {
		return nil, err
	}

	// Store in cache
	i.cache.SetSecret(path, secret)

	return secret, nil
}

// CreateSecret creates a secret and invalidates relevant cache entries
func (i *secretsInteractor) CreateSecret(path string, data map[string]any) error {
	err := i.client.CreateSecret(path, data)
	if err != nil {
		return err
	}

	// Invalidate cache for this path and parent paths
	i.cache.Invalidate(path)
	return nil
}

// UpdateSecret updates a secret and invalidates relevant cache entries
func (i *secretsInteractor) UpdateSecret(path string, data map[string]any) error {
	err := i.client.UpdateSecret(path, data)
	if err != nil {
		return err
	}

	// Invalidate cache for this path and parent paths
	i.cache.Invalidate(path)
	return nil
}

// DeleteSecret deletes a secret and invalidates relevant cache entries
func (i *secretsInteractor) DeleteSecret(path string) error {
	err := i.client.DeleteSecret(path)
	if err != nil {
		return err
	}

	// Invalidate cache for this path and parent paths
	i.cache.Invalidate(path)
	return nil
}

func (i *secretsInteractor) BuildTree(rootPath string, maxDepth int) (*models.SecretNode, error) {
	return i.buildTreeRecursive(rootPath, "", maxDepth, 0)
}

// buildTreeRecursive recursively builds the secret tree
func (i *secretsInteractor) buildTreeRecursive(rootPath, currentPath string, maxDepth, currentDepth int) (*models.SecretNode, error) {
	if currentDepth >= maxDepth {
		return nil, nil
	}

	// List secrets at current path
	secrets, err := i.ListSecrets(currentPath)
	if err != nil {
		return nil, err
	}

	// Create root node
	node := &models.SecretNode{
		Name:     filepath.Base(currentPath),
		Path:     currentPath,
		IsSecret: false,
		Children: []*models.SecretNode{},
	}

	// If this is the root, use the provided name
	if currentPath == "" {
		node.Name = rootPath
		if rootPath == "" {
			node.Name = "secrets"
		}
	}

	// Process each secret/directory
	for _, secret := range secrets {
		if secret.IsSecret {
			// This is a secret, get its data
			secretNode, err := i.GetSecret(secret.Path)
			if err != nil {
				i.logger.Warnf("Failed to get secret '%s': %v", secret.Path, err)
				// Add node without data if we can't retrieve it
				node.Children = append(node.Children, secret)
				continue
			}
			node.Children = append(node.Children, secretNode)
		} else {
			// This is a directory, recurse
			childNode, err := i.buildTreeRecursive(rootPath, secret.Path, maxDepth, currentDepth+1)
			if err != nil {
				i.logger.Warnf("Failed to build tree for path '%s': %v", secret.Path, err)
				continue
			}
			if childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node, nil
}
