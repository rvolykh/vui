package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App   AppConfig   `mapstructure:"app"`
	Vault VaultConfig `mapstructure:"vault"`
	UI    UIConfig    `mapstructure:"ui"`
}

// AppConfig contains general application settings
type AppConfig struct {
	DefaultVault     string `mapstructure:"default_vault"`
	Theme            string `mapstructure:"theme"`
	RefreshInterval  int    `mapstructure:"refresh_interval"`
	MaxSecretSize    int    `mapstructure:"max_secret_size"`
	ClipboardTimeout int    `mapstructure:"clipboard_timeout"`
}

// VaultConfig contains vault connection settings
type VaultConfig struct {
	Address      string                  `mapstructure:"address"`
	Token        string                  `mapstructure:"token"`
	AuthMethod   string                  `mapstructure:"auth_method"`
	Namespace    string                  `mapstructure:"namespace"`
	Profiles     map[string]VaultProfile `mapstructure:"profiles"`
	DefaultVault string                  `mapstructure:"default_vault"`
}

// VaultProfile represents a vault connection profile
type VaultProfile struct {
	Address    string                 `mapstructure:"address"`
	AuthMethod string                 `mapstructure:"auth_method"`
	Token      string                 `mapstructure:"token"`
	Namespace  string                 `mapstructure:"namespace"`
	AuthConfig map[string]interface{} `mapstructure:"auth_config"`
}

// AuthConfig represents authentication configuration for different methods
type AuthConfig struct {
	// Token authentication
	Token string `mapstructure:"token,omitempty"`

	// LDAP authentication
	Username string `mapstructure:"username,omitempty"`
	Password string `mapstructure:"password,omitempty"`

	// AWS authentication
	AWSAccessKeyID     string `mapstructure:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string `mapstructure:"aws_secret_access_key,omitempty"`
	AWSSessionToken    string `mapstructure:"aws_session_token,omitempty"`
	AWSRegion          string `mapstructure:"aws_region,omitempty"`
	AWSRole            string `mapstructure:"aws_role,omitempty"`

	// Azure authentication
	AzureTenantID     string `mapstructure:"azure_tenant_id,omitempty"`
	AzureClientID     string `mapstructure:"azure_client_id,omitempty"`
	AzureClientSecret string `mapstructure:"azure_client_secret,omitempty"`
	AzureResource     string `mapstructure:"azure_resource,omitempty"`

	// GCP authentication
	GCPCredentials string `mapstructure:"gcp_credentials,omitempty"`
	GCPRole        string `mapstructure:"gcp_role,omitempty"`
	GCPProject     string `mapstructure:"gcp_project,omitempty"`

	// Kubernetes authentication
	K8sRole        string `mapstructure:"k8s_role,omitempty"`
	K8sTokenPath   string `mapstructure:"k8s_token_path,omitempty"`
	K8sServicePath string `mapstructure:"k8s_service_path,omitempty"`

	// JWT authentication
	JWTRole string `mapstructure:"jwt_role,omitempty"`
	JWT     string `mapstructure:"jwt,omitempty"`

	// Userpass authentication
	UserpassUsername string `mapstructure:"userpass_username,omitempty"`
	UserpassPassword string `mapstructure:"userpass_password,omitempty"`

	// Cert authentication
	CertName string `mapstructure:"cert_name,omitempty"`
	CertPath string `mapstructure:"cert_path,omitempty"`
	KeyPath  string `mapstructure:"key_path,omitempty"`
}

// UIConfig contains user interface settings
type UIConfig struct {
	ShowHiddenSecrets bool `mapstructure:"show_hidden_secrets"`
	ConfirmDeletions  bool `mapstructure:"confirm_deletions"`
	AutoRefresh       bool `mapstructure:"auto_refresh"`
	TreeWidth         int  `mapstructure:"tree_width"`
}

// Load loads the configuration from files and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("$HOME/.vui")
	viper.AddConfigPath("/etc/vui")

	// Set default values
	setDefaults()

	// Read from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("VUI")

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults and environment variables
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables if set
	overrideWithEnvVars(&config)

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults
	viper.SetDefault("app.default_vault", "default")
	viper.SetDefault("app.theme", "dark")
	viper.SetDefault("app.refresh_interval", 30)
	viper.SetDefault("app.max_secret_size", 10240)
	viper.SetDefault("app.clipboard_timeout", 5)

	// Vault defaults
	viper.SetDefault("vault.address", "http://localhost:8200")
	viper.SetDefault("vault.auth_method", "token")
	viper.SetDefault("vault.namespace", "")

	// UI defaults
	viper.SetDefault("ui.show_hidden_secrets", false)
	viper.SetDefault("ui.confirm_deletions", true)
	viper.SetDefault("ui.auto_refresh", true)
	viper.SetDefault("ui.tree_width", 40)
}

// overrideWithEnvVars overrides configuration with environment variables
func overrideWithEnvVars(config *Config) {
	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		config.Vault.Address = addr
	}
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		config.Vault.Token = token
	}
	if namespace := os.Getenv("VAULT_NAMESPACE"); namespace != "" {
		config.Vault.Namespace = namespace
	}
}

// Save saves the configuration to a file
func (c *Config) Save() error {
	configDir := filepath.Join(os.Getenv("HOME"), ".vui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// Convert config back to viper format for saving
	viper.Set("app", c.App)
	viper.Set("vault", c.Vault)
	viper.Set("ui", c.UI)

	return viper.WriteConfigAs(configFile)
}
