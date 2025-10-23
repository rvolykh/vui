package auth

import (
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

// authenticateWithGCP authenticates using GCP
func (am *AuthManager) authenticateWithGCP(client *api.Client, profile *config.VaultProfile) error {
	role := profile.AuthConfig.GCPRole
	if role == "" {
		return fmt.Errorf("gcp_role is required for GCP authentication")
	}

	project := profile.AuthConfig.GCPProject
	if project == "" {
		return fmt.Errorf("gcp_project is required for GCP authentication")
	}

	// Get GCP credentials
	credentials, err := am.getGCPCredentials(profile)
	if err != nil {
		return fmt.Errorf("failed to get GCP credentials: %w", err)
	}

	// Authenticate with GCP
	secret, err := client.Logical().Write("auth/gcp/login", map[string]interface{}{
		"role":        role,
		"project":     project,
		"credentials": credentials,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with GCP: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from GCP")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}

// getGCPCredentials gets GCP credentials from various sources
func (am *AuthManager) getGCPCredentials(profile *config.VaultProfile) (string, error) {
	// Check if credentials are provided directly
	if credentials := profile.AuthConfig.GCPCredentials; credentials != "" {
		return credentials, nil
	}

	// Check for credentials file path
	if credsPath := profile.AuthConfig.GCPCredentials; credsPath != "" {
		credsData, err := os.ReadFile(credsPath)
		if err != nil {
			return "", fmt.Errorf("failed to read GCP credentials file: %w", err)
		}
		return string(credsData), nil
	}

	// Check for environment variable
	if creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); creds != "" {
		credsData, err := os.ReadFile(creds)
		if err != nil {
			return "", fmt.Errorf("failed to read GCP credentials from environment: %w", err)
		}
		return string(credsData), nil
	}

	return "", fmt.Errorf("no GCP credentials found")
}
