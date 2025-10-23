package vault

import (
	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

type vaultAuthenticator interface {
	VerifyAuthentication(client *api.Client) error
	Authenticate(client *api.Client, profile *config.VaultProfile) error
}

type vaultSecretsManager interface {
	ListSecrets(path string) ([]*SecretNode, error)
	GetSecret(path string) (*SecretNode, error)
	CreateSecret(path string, data map[string]interface{}) error
	UpdateSecret(path string, data map[string]interface{}) error
	DeleteSecret(path string) error
	BuildTree(rootPath string, maxDepth int) (*SecretNode, error)
	SearchSecrets(pattern string, rootPath string) ([]*SecretNode, error)
	SearchSecretsByValue(valuePattern string, rootPath string) ([]*SearchResult, error)
	SearchSecretsByKey(keyPattern string, rootPath string) ([]*SearchResult, error)
}

type vaultConnectionManager interface {
	AddConnection(name string, client *Client)
	TestConnectionAsync(name string)
	RemoveConnection(name string)
	GetConnection(name string) (*Client, error)
	GetConnectionStatus(name string) (*ConnectionStatus, error)
	ListConnections() []string
	RefreshConnectionStatus(name string) error
	RefreshAllConnections()
	SetAllConnecting()
	GetHealthyConnections() []string
	GetConnectedConnections() []string
}
