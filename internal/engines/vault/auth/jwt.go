package auth

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

// authenticateWithJWT authenticates using JWT
func (am *AuthManager) authenticateWithJWT(client *api.Client, profile *config.Profile) error {
	role := profile.AuthConfig.JWTRole
	if role == "" {
		return fmt.Errorf("jwt_role is required for JWT authentication")
	}

	jwt := profile.AuthConfig.JWT
	if jwt == "" {
		return fmt.Errorf("jwt is required for JWT authentication")
	}

	// Authenticate with JWT
	secret, err := client.Logical().Write("auth/jwt/login", map[string]interface{}{
		"role": role,
		"jwt":  jwt,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with JWT: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from JWT")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}
