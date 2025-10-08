package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// VaultProfilesManager manages vault connection profiles
type VaultProfilesManager struct {
	configPath string
	profiles   map[string]*VaultProfile
}

// NewVaultProfilesManager creates a new vault profiles manager
func NewVaultProfilesManager() *VaultProfilesManager {
	configDir := filepath.Join(os.Getenv("HOME"), ".vui")
	return &VaultProfilesManager{
		configPath: filepath.Join(configDir, "vaults.yaml"),
		profiles:   make(map[string]*VaultProfile),
	}
}

// LoadProfiles loads vault profiles from the configuration file
func (vpm *VaultProfilesManager) LoadProfiles() error {
	// Check if config file exists
	if _, err := os.Stat(vpm.configPath); os.IsNotExist(err) {
		// Create default profiles if file doesn't exist
		return vpm.createDefaultProfiles()
	}

	// Load existing profiles
	viper.SetConfigFile(vpm.configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read vault profiles: %w", err)
	}

	// Unmarshal profiles
	var profiles map[string]*VaultProfile
	if err := viper.UnmarshalKey("vaults", &profiles); err != nil {
		return fmt.Errorf("failed to unmarshal vault profiles: %w", err)
	}

	vpm.profiles = profiles
	return nil
}

// SaveProfiles saves vault profiles to the configuration file
func (vpm *VaultProfilesManager) SaveProfiles() error {
	// Ensure config directory exists
	configDir := filepath.Dir(vpm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set up viper for writing
	viper.SetConfigFile(vpm.configPath)
	viper.SetConfigType("yaml")

	// Set the profiles data
	viper.Set("vaults", vpm.profiles)

	// Write to file
	return viper.WriteConfig()
}

// GetProfile returns a vault profile by name
func (vpm *VaultProfilesManager) GetProfile(name string) (*VaultProfile, error) {
	profile, exists := vpm.profiles[name]
	if !exists {
		return nil, fmt.Errorf("vault profile '%s' not found", name)
	}
	return profile, nil
}

// SetProfile sets or updates a vault profile
func (vpm *VaultProfilesManager) SetProfile(name string, profile *VaultProfile) {
	vpm.profiles[name] = profile
}

// DeleteProfile deletes a vault profile
func (vpm *VaultProfilesManager) DeleteProfile(name string) error {
	if _, exists := vpm.profiles[name]; !exists {
		return fmt.Errorf("vault profile '%s' not found", name)
	}
	delete(vpm.profiles, name)
	return nil
}

// ListProfiles returns a list of all profile names
func (vpm *VaultProfilesManager) ListProfiles() []string {
	profiles := make([]string, 0, len(vpm.profiles))
	for name := range vpm.profiles {
		profiles = append(profiles, name)
	}
	return profiles
}

// GetAllProfiles returns all vault profiles
func (vpm *VaultProfilesManager) GetAllProfiles() map[string]*VaultProfile {
	return vpm.profiles
}

// createDefaultProfiles creates default vault profiles
func (vpm *VaultProfilesManager) createDefaultProfiles() error {
	// Create default profile
	defaultProfile := &VaultProfile{
		Address:    "http://localhost:8200",
		AuthMethod: "token",
		Token:      "",
		Namespace:  "",
	}

	vpm.profiles["default"] = defaultProfile

	// Save the default profiles
	return vpm.SaveProfiles()
}

// ValidateProfile validates a vault profile
func (vpm *VaultProfilesManager) ValidateProfile(profile *VaultProfile) error {
	if profile.Address == "" {
		return fmt.Errorf("vault address is required")
	}

	if profile.AuthMethod == "" {
		return fmt.Errorf("auth method is required")
	}

	// Basic validation based on auth method
	switch profile.AuthMethod {
	case "token":
		if profile.Token == "" {
			return fmt.Errorf("token is required for token authentication")
		}
	case "ldap":
		if profile.AuthConfig == nil {
			return fmt.Errorf("auth_config is required for LDAP authentication")
		}
		if username, ok := profile.AuthConfig["username"].(string); !ok || username == "" {
			return fmt.Errorf("username is required for LDAP authentication")
		}
		if password, ok := profile.AuthConfig["password"].(string); !ok || password == "" {
			return fmt.Errorf("password is required for LDAP authentication")
		}
	case "aws":
		if profile.AuthConfig == nil {
			return fmt.Errorf("auth_config is required for AWS authentication")
		}
		if accessKeyID, ok := profile.AuthConfig["aws_access_key_id"].(string); !ok || accessKeyID == "" {
			return fmt.Errorf("aws_access_key_id is required for AWS authentication")
		}
		if secretAccessKey, ok := profile.AuthConfig["aws_secret_access_key"].(string); !ok || secretAccessKey == "" {
			return fmt.Errorf("aws_secret_access_key is required for AWS authentication")
		}
		if role, ok := profile.AuthConfig["aws_role"].(string); !ok || role == "" {
			return fmt.Errorf("aws_role is required for AWS authentication")
		}
	case "azure":
		if profile.AuthConfig == nil {
			return fmt.Errorf("auth_config is required for Azure authentication")
		}
		if tenantID, ok := profile.AuthConfig["azure_tenant_id"].(string); !ok || tenantID == "" {
			return fmt.Errorf("azure_tenant_id is required for Azure authentication")
		}
		if clientID, ok := profile.AuthConfig["azure_client_id"].(string); !ok || clientID == "" {
			return fmt.Errorf("azure_client_id is required for Azure authentication")
		}
		if clientSecret, ok := profile.AuthConfig["azure_client_secret"].(string); !ok || clientSecret == "" {
			return fmt.Errorf("azure_client_secret is required for Azure authentication")
		}
	case "gcp":
		if profile.AuthConfig == nil {
			return fmt.Errorf("auth_config is required for GCP authentication")
		}
		if role, ok := profile.AuthConfig["gcp_role"].(string); !ok || role == "" {
			return fmt.Errorf("gcp_role is required for GCP authentication")
		}
	case "kubernetes":
		if profile.AuthConfig == nil {
			return fmt.Errorf("auth_config is required for Kubernetes authentication")
		}
		if role, ok := profile.AuthConfig["k8s_role"].(string); !ok || role == "" {
			return fmt.Errorf("k8s_role is required for Kubernetes authentication")
		}
	case "jwt":
		if profile.AuthConfig == nil {
			return fmt.Errorf("auth_config is required for JWT authentication")
		}
		if role, ok := profile.AuthConfig["jwt_role"].(string); !ok || role == "" {
			return fmt.Errorf("jwt_role is required for JWT authentication")
		}
		if jwt, ok := profile.AuthConfig["jwt"].(string); !ok || jwt == "" {
			return fmt.Errorf("jwt is required for JWT authentication")
		}
	case "userpass":
		if profile.AuthConfig == nil {
			return fmt.Errorf("auth_config is required for userpass authentication")
		}
		if username, ok := profile.AuthConfig["userpass_username"].(string); !ok || username == "" {
			return fmt.Errorf("userpass_username is required for userpass authentication")
		}
		if password, ok := profile.AuthConfig["userpass_password"].(string); !ok || password == "" {
			return fmt.Errorf("userpass_password is required for userpass authentication")
		}
	case "cert":
		if profile.AuthConfig == nil {
			return fmt.Errorf("auth_config is required for cert authentication")
		}
		if certName, ok := profile.AuthConfig["cert_name"].(string); !ok || certName == "" {
			return fmt.Errorf("cert_name is required for cert authentication")
		}
		if certPath, ok := profile.AuthConfig["cert_path"].(string); !ok || certPath == "" {
			return fmt.Errorf("cert_path is required for cert authentication")
		}
		if keyPath, ok := profile.AuthConfig["key_path"].(string); !ok || keyPath == "" {
			return fmt.Errorf("key_path is required for cert authentication")
		}
	default:
		return fmt.Errorf("unsupported auth method: %s", profile.AuthMethod)
	}

	return nil
}

// GetConfigPath returns the configuration file path
func (vpm *VaultProfilesManager) GetConfigPath() string {
	return vpm.configPath
}
