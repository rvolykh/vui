package auth

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

// authenticateWithToken authenticates using a token
func (am *AuthManager) authenticateWithToken(client *api.Client, profile *config.VaultProfile) error {
	token := profile.AuthConfig.Token
	if token == "" {
		return fmt.Errorf("token is required for token authentication")
	}

	client.SetToken(token)
	return nil
}
