package auth

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

// authenticateWithUserpass authenticates using userpass
func (am *AuthManager) authenticateWithUserpass(client *api.Client, profile *config.Profile) error {
	username := profile.AuthConfig.Username
	if username == "" {
		return fmt.Errorf("username is required for userpass authentication")
	}

	password := profile.AuthConfig.Password
	if password == "" {
		return fmt.Errorf("password is required for userpass authentication")
	}

	// Authenticate with userpass
	secret, err := client.Logical().Write(fmt.Sprintf("auth/userpass/login/%s", username), map[string]interface{}{
		"password": password,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with userpass: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from userpass")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}
