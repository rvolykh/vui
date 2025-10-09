package vault

import (
	"fmt"
	"io/ioutil"
	"os"

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

// authenticateWithToken authenticates using a token
func (am *AuthManager) authenticateWithToken(client *api.Client, profile *config.VaultProfile) error {
	if profile.Token == "" {
		return fmt.Errorf("token is required for token authentication")
	}

	client.SetToken(profile.Token)
	return nil
}

// authenticateWithLDAP authenticates using LDAP
func (am *AuthManager) authenticateWithLDAP(client *api.Client, profile *config.VaultProfile) error {
	username, ok := profile.AuthConfig["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("username is required for LDAP authentication")
	}

	password, ok := profile.AuthConfig["password"].(string)
	if !ok || password == "" {
		return fmt.Errorf("password is required for LDAP authentication")
	}

	// Authenticate with LDAP
	secret, err := client.Logical().Write("auth/ldap/login", map[string]interface{}{
		"username": username,
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

// authenticateWithAWS authenticates using AWS
func (am *AuthManager) authenticateWithAWS(client *api.Client, profile *config.VaultProfile) error {
	accessKeyID, ok := profile.AuthConfig["aws_access_key_id"].(string)
	if !ok || accessKeyID == "" {
		return fmt.Errorf("aws_access_key_id is required for AWS authentication")
	}

	secretAccessKey, ok := profile.AuthConfig["aws_secret_access_key"].(string)
	if !ok || secretAccessKey == "" {
		return fmt.Errorf("aws_secret_access_key is required for AWS authentication")
	}

	role, ok := profile.AuthConfig["aws_role"].(string)
	if !ok || role == "" {
		return fmt.Errorf("aws_role is required for AWS authentication")
	}

	region, _ := profile.AuthConfig["aws_region"].(string)
	if region == "" {
		region = "us-east-1"
	}

	sessionToken, _ := profile.AuthConfig["aws_session_token"].(string)

	// Authenticate with AWS
	authData := map[string]interface{}{
		"role":                    role,
		"iam_http_request_method": "POST",
		"iam_request_url":         "https://sts.amazonaws.com/",
		"iam_request_body":        "Action=GetCallerIdentity&Version=2011-06-15",
		"iam_request_headers": map[string]string{
			"Content-Type": "application/x-www-form-urlencoded; charset=utf-8",
		},
	}

	if sessionToken != "" {
		authData["iam_request_headers"].(map[string]string)["X-Amz-Security-Token"] = sessionToken
	}

	secret, err := client.Logical().Write("auth/aws/login", authData)
	if err != nil {
		return fmt.Errorf("failed to authenticate with AWS: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from AWS")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateWithAzure authenticates using Azure
func (am *AuthManager) authenticateWithAzure(client *api.Client, profile *config.VaultProfile) error {
	tenantID, ok := profile.AuthConfig["azure_tenant_id"].(string)
	if !ok || tenantID == "" {
		return fmt.Errorf("azure_tenant_id is required for Azure authentication")
	}

	clientID, ok := profile.AuthConfig["azure_client_id"].(string)
	if !ok || clientID == "" {
		return fmt.Errorf("azure_client_id is required for Azure authentication")
	}

	clientSecret, ok := profile.AuthConfig["azure_client_secret"].(string)
	if !ok || clientSecret == "" {
		return fmt.Errorf("azure_client_secret is required for Azure authentication")
	}

	resource, _ := profile.AuthConfig["azure_resource"].(string)
	if resource == "" {
		resource = "https://management.azure.com/"
	}

	// Authenticate with Azure
	secret, err := client.Logical().Write("auth/azure/login", map[string]interface{}{
		"role":          profile.AuthConfig["azure_role"],
		"tenant_id":     tenantID,
		"client_id":     clientID,
		"client_secret": clientSecret,
		"resource":      resource,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with Azure: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from Azure")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateWithGCP authenticates using GCP
func (am *AuthManager) authenticateWithGCP(client *api.Client, profile *config.VaultProfile) error {
	role, ok := profile.AuthConfig["gcp_role"].(string)
	if !ok || role == "" {
		return fmt.Errorf("gcp_role is required for GCP authentication")
	}

	project, _ := profile.AuthConfig["gcp_project"].(string)

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

// authenticateWithKubernetes authenticates using Kubernetes
func (am *AuthManager) authenticateWithKubernetes(client *api.Client, profile *config.VaultProfile) error {
	role, ok := profile.AuthConfig["k8s_role"].(string)
	if !ok || role == "" {
		return fmt.Errorf("k8s_role is required for Kubernetes authentication")
	}

	// Get Kubernetes token
	token, err := am.getKubernetesToken(profile)
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes token: %w", err)
	}

	// Authenticate with Kubernetes
	secret, err := client.Logical().Write("auth/kubernetes/login", map[string]interface{}{
		"role": role,
		"jwt":  token,
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with Kubernetes: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from Kubernetes")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateWithJWT authenticates using JWT
func (am *AuthManager) authenticateWithJWT(client *api.Client, profile *config.VaultProfile) error {
	role, ok := profile.AuthConfig["jwt_role"].(string)
	if !ok || role == "" {
		return fmt.Errorf("jwt_role is required for JWT authentication")
	}

	jwt, ok := profile.AuthConfig["jwt"].(string)
	if !ok || jwt == "" {
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

// authenticateWithUserpass authenticates using userpass
func (am *AuthManager) authenticateWithUserpass(client *api.Client, profile *config.VaultProfile) error {
	username, ok := profile.AuthConfig["userpass_username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("userpass_username is required for userpass authentication")
	}

	password, ok := profile.AuthConfig["userpass_password"].(string)
	if !ok || password == "" {
		return fmt.Errorf("userpass_password is required for userpass authentication")
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

// authenticateWithCert authenticates using certificates
func (am *AuthManager) authenticateWithCert(client *api.Client, profile *config.VaultProfile) error {
	certName, ok := profile.AuthConfig["cert_name"].(string)
	if !ok || certName == "" {
		return fmt.Errorf("cert_name is required for cert authentication")
	}

	certPath, ok := profile.AuthConfig["cert_path"].(string)
	if !ok || certPath == "" {
		return fmt.Errorf("cert_path is required for cert authentication")
	}

	keyPath, ok := profile.AuthConfig["key_path"].(string)
	if !ok || keyPath == "" {
		return fmt.Errorf("key_path is required for cert authentication")
	}

	// Read certificate and key files (Note: This is a simplified implementation)
	// In a real implementation, you would need to configure TLS properly
	_, err := ioutil.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	_, err = ioutil.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Set client certificate (Note: This is a simplified implementation)
	// In a real implementation, you would need to configure TLS properly
	// For now, we'll just validate that the files exist

	// Authenticate with cert
	secret, err := client.Logical().Write("auth/cert/login", map[string]interface{}{
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

// getGCPCredentials gets GCP credentials from various sources
func (am *AuthManager) getGCPCredentials(profile *config.VaultProfile) (string, error) {
	// Check if credentials are provided directly
	if credentials, ok := profile.AuthConfig["gcp_credentials"].(string); ok && credentials != "" {
		return credentials, nil
	}

	// Check for credentials file path
	if credsPath, ok := profile.AuthConfig["gcp_credentials_path"].(string); ok && credsPath != "" {
		credsData, err := ioutil.ReadFile(credsPath)
		if err != nil {
			return "", fmt.Errorf("failed to read GCP credentials file: %w", err)
		}
		return string(credsData), nil
	}

	// Check for environment variable
	if creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); creds != "" {
		credsData, err := ioutil.ReadFile(creds)
		if err != nil {
			return "", fmt.Errorf("failed to read GCP credentials from environment: %w", err)
		}
		return string(credsData), nil
	}

	return "", fmt.Errorf("no GCP credentials found")
}

// getKubernetesToken gets Kubernetes service account token
func (am *AuthManager) getKubernetesToken(profile *config.VaultProfile) (string, error) {
	// Check if token is provided directly
	if token, ok := profile.AuthConfig["k8s_token"].(string); ok && token != "" {
		return token, nil
	}

	// Check for custom token path
	tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	if customPath, ok := profile.AuthConfig["k8s_token_path"].(string); ok && customPath != "" {
		tokenPath = customPath
	}

	// Read token from file
	tokenData, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Kubernetes token: %w", err)
	}

	return string(tokenData), nil
}

// ValidateAuthConfig validates authentication configuration for a profile
func (am *AuthManager) ValidateAuthConfig(profile *config.VaultProfile) error {
	switch profile.AuthMethod {
	case "token":
		return am.validateTokenConfig(profile)
	case "ldap":
		return am.validateLDAPConfig(profile)
	case "aws":
		return am.validateAWSConfig(profile)
	case "azure":
		return am.validateAzureConfig(profile)
	case "gcp":
		return am.validateGCPConfig(profile)
	case "kubernetes":
		return am.validateKubernetesConfig(profile)
	case "jwt":
		return am.validateJWTConfig(profile)
	case "userpass":
		return am.validateUserpassConfig(profile)
	case "cert":
		return am.validateCertConfig(profile)
	default:
		return fmt.Errorf("unsupported authentication method: %s", profile.AuthMethod)
	}
}

// validateTokenConfig validates token authentication configuration
func (am *AuthManager) validateTokenConfig(profile *config.VaultProfile) error {
	if profile.Token == "" {
		return fmt.Errorf("token is required for token authentication")
	}
	return nil
}

// validateLDAPConfig validates LDAP authentication configuration
func (am *AuthManager) validateLDAPConfig(profile *config.VaultProfile) error {
	if username, ok := profile.AuthConfig["username"].(string); !ok || username == "" {
		return fmt.Errorf("username is required for LDAP authentication")
	}
	if password, ok := profile.AuthConfig["password"].(string); !ok || password == "" {
		return fmt.Errorf("password is required for LDAP authentication")
	}
	return nil
}

// validateAWSConfig validates AWS authentication configuration
func (am *AuthManager) validateAWSConfig(profile *config.VaultProfile) error {
	if accessKeyID, ok := profile.AuthConfig["aws_access_key_id"].(string); !ok || accessKeyID == "" {
		return fmt.Errorf("aws_access_key_id is required for AWS authentication")
	}
	if secretAccessKey, ok := profile.AuthConfig["aws_secret_access_key"].(string); !ok || secretAccessKey == "" {
		return fmt.Errorf("aws_secret_access_key is required for AWS authentication")
	}
	if role, ok := profile.AuthConfig["aws_role"].(string); !ok || role == "" {
		return fmt.Errorf("aws_role is required for AWS authentication")
	}
	return nil
}

// validateAzureConfig validates Azure authentication configuration
func (am *AuthManager) validateAzureConfig(profile *config.VaultProfile) error {
	if tenantID, ok := profile.AuthConfig["azure_tenant_id"].(string); !ok || tenantID == "" {
		return fmt.Errorf("azure_tenant_id is required for Azure authentication")
	}
	if clientID, ok := profile.AuthConfig["azure_client_id"].(string); !ok || clientID == "" {
		return fmt.Errorf("azure_client_id is required for Azure authentication")
	}
	if clientSecret, ok := profile.AuthConfig["azure_client_secret"].(string); !ok || clientSecret == "" {
		return fmt.Errorf("azure_client_secret is required for Azure authentication")
	}
	return nil
}

// validateGCPConfig validates GCP authentication configuration
func (am *AuthManager) validateGCPConfig(profile *config.VaultProfile) error {
	if role, ok := profile.AuthConfig["gcp_role"].(string); !ok || role == "" {
		return fmt.Errorf("gcp_role is required for GCP authentication")
	}
	return nil
}

// validateKubernetesConfig validates Kubernetes authentication configuration
func (am *AuthManager) validateKubernetesConfig(profile *config.VaultProfile) error {
	if role, ok := profile.AuthConfig["k8s_role"].(string); !ok || role == "" {
		return fmt.Errorf("k8s_role is required for Kubernetes authentication")
	}
	return nil
}

// validateJWTConfig validates JWT authentication configuration
func (am *AuthManager) validateJWTConfig(profile *config.VaultProfile) error {
	if role, ok := profile.AuthConfig["jwt_role"].(string); !ok || role == "" {
		return fmt.Errorf("jwt_role is required for JWT authentication")
	}
	if jwt, ok := profile.AuthConfig["jwt"].(string); !ok || jwt == "" {
		return fmt.Errorf("jwt is required for JWT authentication")
	}
	return nil
}

// validateUserpassConfig validates userpass authentication configuration
func (am *AuthManager) validateUserpassConfig(profile *config.VaultProfile) error {
	if username, ok := profile.AuthConfig["userpass_username"].(string); !ok || username == "" {
		return fmt.Errorf("userpass_username is required for userpass authentication")
	}
	if password, ok := profile.AuthConfig["userpass_password"].(string); !ok || password == "" {
		return fmt.Errorf("userpass_password is required for userpass authentication")
	}
	return nil
}

// validateCertConfig validates certificate authentication configuration
func (am *AuthManager) validateCertConfig(profile *config.VaultProfile) error {
	if certName, ok := profile.AuthConfig["cert_name"].(string); !ok || certName == "" {
		return fmt.Errorf("cert_name is required for cert authentication")
	}
	certPath, ok := profile.AuthConfig["cert_path"].(string)
	if !ok || certPath == "" {
		return fmt.Errorf("cert_path is required for cert authentication")
	}
	keyPath, ok := profile.AuthConfig["key_path"].(string)
	if !ok || keyPath == "" {
		return fmt.Errorf("key_path is required for cert authentication")
	}

	// Check if files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("certificate file does not exist: %s", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("key file does not exist: %s", keyPath)
	}

	return nil
}
