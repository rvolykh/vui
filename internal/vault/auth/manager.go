package auth

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
)

// AuthManager manages authentication for different vault connections
type AuthManager struct {
	logger *logrus.Logger
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(logger *logrus.Logger) *AuthManager {
	return &AuthManager{
		logger: logger,
	}
}

// VerifyAuthentication verifies that the client is authenticated by checking the token
func (am *AuthManager) VerifyAuthentication(client *api.Client) error {
	// Try to lookup the current token
	secret, err := client.Auth().Token().LookupSelf()
	if err != nil {
		return fmt.Errorf("authentication verification failed: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return fmt.Errorf("no token data returned")
	}

	am.logger.Debug("Authentication verified successfully")
	return nil
}

// Authenticate authenticates a client using the specified profile
func (am *AuthManager) Authenticate(client *api.Client, profile *config.VaultProfile) error {
	switch profile.AuthMethod {
	case "token":
		return am.authenticateWithToken(client, profile)
	case "ldap":
		return am.authenticateWithLDAP(client, profile)
	case "aws":
		return am.authenticateWithAWS(client, profile)
	case "azure":
		return am.authenticateWithAzure(client, profile)
	case "gcp":
		return am.authenticateWithGCP(client, profile)
	case "kubernetes":
		return am.authenticateWithKubernetes(client, profile)
	case "oidc":
		return am.authenticateWithOIDC(client, profile)
	case "jwt":
		return am.authenticateWithJWT(client, profile)
	case "userpass":
		return am.authenticateWithUserpass(client, profile)
	case "cert":
		return am.authenticateWithCert(client, profile)
	default:
		return fmt.Errorf("unsupported authentication method: %s", profile.AuthMethod)
	}
}
