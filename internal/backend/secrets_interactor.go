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
	engines.SecretClient
}

func newSecretsInteractor(logger *logrus.Logger, name string, client engines.SecretClient) SecretsInteractor {
	return &secretsInteractor{
		logger:       logger,
		name:         name,
		SecretClient: client,
	}
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
