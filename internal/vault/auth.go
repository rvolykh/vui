package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/go-secure-stdlib/awsutil"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/azure"
	"github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/rvolykh/vui/internal/adapters"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/utils"
	"github.com/sirupsen/logrus"
	k8sauth "k8s.io/api/authentication/v1"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	k8scmd "k8s.io/client-go/tools/clientcmd"
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
	token := profile.AuthConfig.Token
	if token == "" {
		return fmt.Errorf("token is required for token authentication")
	}

	client.SetToken(token)
	return nil
}

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

// authenticateWithAWS authenticates using AWS
func (am *AuthManager) authenticateWithAWS(client *api.Client, profile *config.VaultProfile) error {
	accessKeyID := profile.AuthConfig.AWSAccessKeyID
	if accessKeyID == "" {
		return fmt.Errorf("aws_access_key_id is required for AWS authentication")
	}

	secretAccessKey := profile.AuthConfig.AWSSecretAccessKey
	if secretAccessKey == "" {
		return fmt.Errorf("aws_secret_access_key is required for AWS authentication")
	}

	sessionToken := profile.AuthConfig.AWSSessionToken

	role := profile.AuthConfig.AWSRole
	if role == "" {
		return fmt.Errorf("aws_role is required for AWS authentication")
	}

	region := profile.AuthConfig.AWSRegion
	if region == "" {
		region = "us-east-1"
	}

	creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, sessionToken)

	data, err := awsutil.GenerateLoginData(creds, "", region, adapters.NewHclogAdapter(am.logger))
	if err != nil {
		return fmt.Errorf("unable to generate login data for AWS auth endpoint: %w", err)
	}

	secret, err := client.Logical().Write("auth/aws/login", data)
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

// authenticateWithKubernetes authenticates using Kubernetes
func (am *AuthManager) authenticateWithKubernetes(client *api.Client, profile *config.VaultProfile) error {
	role := profile.AuthConfig.K8sRole
	if role == "" {
		return fmt.Errorf("k8s_role is required for Kubernetes authentication")
	}

	var (
		token string
		err   error

		opts []kubernetes.LoginOption
	)
	if profile.AuthConfig.K8sToken != "" {
		token = profile.AuthConfig.K8sToken
	} else if profile.AuthConfig.K8sServiceAccount != "" {
		token, err = am.getKubernetesToken(profile)
		if err != nil {
			return fmt.Errorf("failed to get Kubernetes token: %w", err)
		}
	}
	if token != "" {
		opts = append(opts, kubernetes.WithServiceAccountToken(token))
	}

	k8sAuth, err := kubernetes.NewKubernetesAuth(role, opts...)
	if err != nil {
		return fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.TODO(), k8sAuth)
	if err != nil {
		return fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
	}
	if authInfo == nil {
		return fmt.Errorf("no auth info was returned after login")
	}

	return nil
}

// authenticateWithJWT authenticates using JWT
func (am *AuthManager) authenticateWithJWT(client *api.Client, profile *config.VaultProfile) error {
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

// authenticateWithUserpass authenticates using userpass
func (am *AuthManager) authenticateWithUserpass(client *api.Client, profile *config.VaultProfile) error {
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

// authenticateWithCert authenticates using certificates
func (am *AuthManager) authenticateWithCert(client *api.Client, profile *config.VaultProfile) error {
	certName := profile.AuthConfig.CertName
	if certName == "" {
		return fmt.Errorf("cert_name is required for cert authentication")
	}

	certPath := profile.AuthConfig.CertPath
	if certPath == "" {
		return fmt.Errorf("cert_path is required for cert authentication")
	}

	keyPath := profile.AuthConfig.KeyPath
	if keyPath == "" {
		return fmt.Errorf("key_path is required for cert authentication")
	}

	// Read certificate and key files (Note: This is a simplified implementation)
	// In a real implementation, you would need to configure TLS properly
	_, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	_, err = os.ReadFile(keyPath)
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

// getKubernetesToken gets Kubernetes service account token
func (am *AuthManager) getKubernetesToken(profile *config.VaultProfile) (string, error) {
	homeDir, _ := os.UserHomeDir()

	var (
		serviceAccount = profile.AuthConfig.K8sServiceAccount

		kubeconfigPath = utils.Coalesce(profile.AuthConfig.K8sConfigPath, filepath.Join(homeDir, ".kube", "config"))
		audience       = utils.Coalesce(profile.AuthConfig.K8sAudience, "vault")
		namespace      = utils.Coalesce(profile.AuthConfig.K8sNamespace, "default")
		ttl            = utils.Coalesce(profile.AuthConfig.K8sTTL, int64(3600))
	)

	config, err := k8scmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to build Kubernetes config: %w", err)
	}

	clientset, err := k8s.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	tr := &k8sauth.TokenRequest{
		Spec: k8sauth.TokenRequestSpec{
			Audiences:         []string{audience},
			ExpirationSeconds: &ttl,
		},
	}
	tokenRequest, err := clientset.CoreV1().ServiceAccounts(namespace).
		CreateToken(context.TODO(), serviceAccount, tr, k8smeta.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes token: %w", err)
	}

	return tokenRequest.Status.Token, nil
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
	if token := profile.AuthConfig.Token; token == "" {
		return fmt.Errorf("token is required for token authentication")
	}
	return nil
}

// validateLDAPConfig validates LDAP authentication configuration
func (am *AuthManager) validateLDAPConfig(profile *config.VaultProfile) error {
	if username := profile.AuthConfig.Username; username == "" {
		return fmt.Errorf("username is required for LDAP authentication")
	}
	if password := profile.AuthConfig.Password; password == "" {
		return fmt.Errorf("password is required for LDAP authentication")
	}
	return nil
}

// validateAWSConfig validates AWS authentication configuration
func (am *AuthManager) validateAWSConfig(profile *config.VaultProfile) error {
	if accessKeyID := profile.AuthConfig.AWSAccessKeyID; accessKeyID == "" {
		return fmt.Errorf("aws_access_key_id is required for AWS authentication")
	}
	if secretAccessKey := profile.AuthConfig.AWSSecretAccessKey; secretAccessKey == "" {
		return fmt.Errorf("aws_secret_access_key is required for AWS authentication")
	}
	if role := profile.AuthConfig.AWSRole; role == "" {
		return fmt.Errorf("aws_role is required for AWS authentication")
	}
	return nil
}

// validateAzureConfig validates Azure authentication configuration
func (am *AuthManager) validateAzureConfig(profile *config.VaultProfile) error {
	if role := profile.AuthConfig.AzureRole; role == "" {
		return fmt.Errorf("azure_role is required for Azure authentication")
	}
	return nil
}

// validateGCPConfig validates GCP authentication configuration
func (am *AuthManager) validateGCPConfig(profile *config.VaultProfile) error {
	if role := profile.AuthConfig.GCPRole; role == "" {
		return fmt.Errorf("gcp_role is required for GCP authentication")
	}
	return nil
}

// validateKubernetesConfig validates Kubernetes authentication configuration
func (am *AuthManager) validateKubernetesConfig(profile *config.VaultProfile) error {
	if role := profile.AuthConfig.K8sRole; role == "" {
		return fmt.Errorf("k8s_role is required for Kubernetes authentication")
	}
	return nil
}

// validateJWTConfig validates JWT authentication configuration
func (am *AuthManager) validateJWTConfig(profile *config.VaultProfile) error {
	if role := profile.AuthConfig.JWTRole; role == "" {
		return fmt.Errorf("jwt_role is required for JWT authentication")
	}
	if jwt := profile.AuthConfig.JWT; jwt == "" {
		return fmt.Errorf("jwt is required for JWT authentication")
	}
	return nil
}

// validateUserpassConfig validates userpass authentication configuration
func (am *AuthManager) validateUserpassConfig(profile *config.VaultProfile) error {
	if username := profile.AuthConfig.Username; username == "" {
		return fmt.Errorf("username is required for userpass authentication")
	}
	if password := profile.AuthConfig.Password; password == "" {
		return fmt.Errorf("password is required for userpass authentication")
	}
	return nil
}

// validateCertConfig validates certificate authentication configuration
func (am *AuthManager) validateCertConfig(profile *config.VaultProfile) error {
	if certName := profile.AuthConfig.CertName; certName == "" {
		return fmt.Errorf("cert_name is required for cert authentication")
	}
	certPath := profile.AuthConfig.CertPath
	if certPath == "" {
		return fmt.Errorf("cert_path is required for cert authentication")
	}
	keyPath := profile.AuthConfig.KeyPath
	if keyPath == "" {
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
