package auth

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

// authenticateWithLDAP authenticates using LDAP
func (am *AuthManager) authenticateWithLDAP(client *api.Client, profile *config.VaultProfile) error {
	username := profile.AuthConfig.Username
	if username == "" {
		return fmt.Errorf("username is required for LDAP authentication")
	}

	password := profile.AuthConfig.Password
	if password == "" {
		return fmt.Errorf("password is required for LDAP authentication")
	}

	// Authenticate with LDAP
	secret, err := client.Logical().Write("auth/ldap/login/"+username, map[string]interface{}{
		"password": password,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with LDAP: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from LDAP")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}
