package auth

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/azure"
	"github.com/rvolykh/vui/internal/config"
)

// authenticateWithAzure authenticates using Azure
func (am *AuthManager) authenticateWithAzure(client *api.Client, profile *config.VaultProfile) error {
	role := profile.AuthConfig.AzureRole
	if role == "" {
		return fmt.Errorf("azure_role is required for Azure authentication")
	}

	opts := make([]azure.LoginOption, 0)
	resource := profile.AuthConfig.AzureResource
	if resource != "" {
		opts = append(opts, azure.WithResource(resource))
	}

	azAuth, err := azure.NewAzureAuth(role, opts...)
	if err != nil {
		return fmt.Errorf("unable to initialize Azure auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.TODO(), azAuth)
	if err != nil {
		return fmt.Errorf("unable to login to Azure auth method: %w", err)
	}
	if authInfo == nil {
		return fmt.Errorf("no auth info was returned after login")
	}

	return nil
}
