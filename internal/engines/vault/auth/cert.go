package auth

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

// authenticateWithCert authenticates using certificates
func (am *AuthManager) authenticateWithCert(client *api.Client, profile *config.Profile) error {
	certName := profile.AuthConfig.CertName
	if certName == "" {
		return fmt.Errorf("cert_name is required for cert authentication")
	}

	certPath := profile.AuthConfig.CertCrtPath
	if certPath == "" {
		return fmt.Errorf("cert_path is required for cert authentication")
	}

	keyPath := profile.AuthConfig.CertKeyPath
	if keyPath == "" {
		return fmt.Errorf("key_path is required for cert authentication")
	}

	cfg := api.DefaultConfig()
	cfg.Address = client.Address()

	if err := cfg.ConfigureTLS(&api.TLSConfig{
		CACert:     profile.CertPath,
		ClientCert: certPath,
		ClientKey:  keyPath,
	}); err != nil {
		return fmt.Errorf("failed to configure client TLS: %w", err)
	}

	certClient, err := api.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create new client: %w", err)
	}

	if ns := client.Namespace(); ns != "" {
		certClient.SetNamespace(ns)
	}

	secret, err := certClient.Logical().Write("auth/cert/login", map[string]interface{}{
		"name": certName,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with cert: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from cert")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}
