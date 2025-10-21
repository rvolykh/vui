package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App    AppConfig               `mapstructure:"app"`
	UI     UIConfig                `mapstructure:"ui"`
	Vaults map[string]VaultProfile `mapstructure:"vaults"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Theme            string `mapstructure:"theme"`
	RefreshInterval  int    `mapstructure:"refresh_interval"`
	MaxSecretSize    int    `mapstructure:"max_secret_size"`
	ClipboardTimeout int    `mapstructure:"clipboard_timeout"`
}

// VaultProfile represents a vault connection profile
type VaultProfile struct {
	Address    string     `mapstructure:"address"`
	CertPath   string     `mapstructure:"cert_path,omitempty"`
	AuthMethod string     `mapstructure:"auth_method"`
	Namespace  string     `mapstructure:"namespace"`
	AuthConfig AuthConfig `mapstructure:"auth_config"`
}

// AuthConfig represents authentication configuration for different methods
type AuthConfig struct {
	// Token authentication
	Token string `mapstructure:"token,omitempty"`

	// LDAP authentication & Userpass authentication
	Username string `mapstructure:"username,omitempty"`
	Password string `mapstructure:"password,omitempty"`

	// AWS authentication
	AWSRole            string `mapstructure:"aws_role,omitempty"`
	AWSAccessKeyID     string `mapstructure:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string `mapstructure:"aws_secret_access_key,omitempty"`
	AWSSessionToken    string `mapstructure:"aws_session_token,omitempty"`
	AWSRegion          string `mapstructure:"aws_region,omitempty"`

	// Azure authentication
	AzureRole     string `mapstructure:"azure_role,omitempty"`
	AzureResource string `mapstructure:"azure_resource,omitempty"`

	// GCP authentication
	GCPRole        string `mapstructure:"gcp_role,omitempty"`
	GCPCredentials string `mapstructure:"gcp_credentials,omitempty"`
	GCPProject     string `mapstructure:"gcp_project,omitempty"`

	// Kubernetes authentication
	K8sRole           string `mapstructure:"k8s_role,omitempty"`
	K8sToken          string `mapstructure:"k8s_token,omitempty"`
	K8sConfigPath     string `mapstructure:"k8s_config_path,omitempty"`
	K8sNamespace      string `mapstructure:"k8s_namespace,omitempty"`
	K8sServiceAccount string `mapstructure:"k8s_service_account,omitempty"`
	K8sAudience       string `mapstructure:"k8s_audience,omitempty"`
	K8sTTL            int64  `mapstructure:"k8s_ttl,omitempty"`

	// OIDC authentication
	OIDCRole string `mapstructure:"oidc_role,omitempty"`

	// JWT authentication
	JWTRole string `mapstructure:"jwt_role,omitempty"`
	JWT     string `mapstructure:"jwt,omitempty"`

	// Cert authentication
	CertName    string `mapstructure:"cert_name,omitempty"`
	CertCrtPath string `mapstructure:"cert_crt_path,omitempty"`
	CertKeyPath string `mapstructure:"cert_key_path,omitempty"`
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
	viper.SetConfigName("vui.yaml")
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

		// Create config file
		os.MkdirAll(filepath.Join(os.Getenv("HOME"), ".vui"), 0755)
		configFile := filepath.Join(os.Getenv("HOME"), ".vui", "vui.yaml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			viper.WriteConfigAs(configFile)
		}
	}

	// Expand environment variables in config
	for _, k := range viper.AllKeys() {
		v := viper.GetString(k)
		viper.Set(k, os.ExpandEnv(v))
	}

	// Parse config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults
	viper.SetDefault("app.theme", "dark")
	viper.SetDefault("app.refresh_interval", 30)
	viper.SetDefault("app.max_secret_size", 10240)
	viper.SetDefault("app.clipboard_timeout", 5)

	// UI defaults
	viper.SetDefault("ui.show_hidden_secrets", false)
	viper.SetDefault("ui.confirm_deletions", true)
	viper.SetDefault("ui.auto_refresh", true)
	viper.SetDefault("ui.tree_width", 40)

	// Vaults defaults
	viper.SetDefault("vaults.local.address", "http://localhost:8200")
	viper.SetDefault("vaults.local.auth_method", "token")
	viper.SetDefault("vaults.local.namespace", "")
	viper.SetDefault("vaults.local.auth_config.token", "${VAULT_TOKEN}")
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
	viper.Set("ui", c.UI)
	viper.Set("vaults", c.Vaults)

	return viper.WriteConfigAs(configFile)
}
