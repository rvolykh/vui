package vault

import (
	"fmt"
	"sync"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
)

// Manager manages multiple Vault connections
type Manager struct {
	config        *config.VaultConfig
	clients       map[string]*Client
	activeVault   string
	connectionMgr *ConnectionManager
	secretsMgr    *SecretsManager
	profilesMgr   *config.VaultProfilesManager
	authMgr       *AuthManager
	mutex         sync.RWMutex
	logger        *logrus.Logger
}

// Client wraps the Vault API client with additional functionality
type Client struct {
	apiClient *api.Client
	config    *config.VaultConfig
	logger    *logrus.Logger
}

// NewManager creates a new vault manager
func NewManager(cfg *config.VaultConfig) (*Manager, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Initialize profiles manager
	profilesMgr := config.NewVaultProfilesManager()
	if err := profilesMgr.LoadProfiles(); err != nil {
		logger.Warnf("Failed to load vault profiles: %v", err)
	}

	manager := &Manager{
		config:        cfg,
		clients:       make(map[string]*Client),
		activeVault:   cfg.DefaultVault,
		connectionMgr: NewConnectionManager(logger),
		profilesMgr:   profilesMgr,
		authMgr:       NewAuthManager(logger),
		logger:        logger,
	}

	// Initialize all configured vault clients
	if err := manager.initializeAllClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize vault clients: %w", err)
	}

	return manager, nil
}

// initializeAllClients initializes all configured vault clients
func (m *Manager) initializeAllClients() error {
	// Initialize default client first
	if err := m.initializeDefaultClient(); err != nil {
		m.logger.Warnf("Failed to initialize default client: %v", err)
	}

	// Initialize all profile clients
	profiles := m.profilesMgr.GetAllProfiles()
	for name, profile := range profiles {
		if name == "default" {
			continue // Already initialized
		}

		if err := m.initializeProfileClient(name, profile); err != nil {
			m.logger.Warnf("Failed to initialize client for profile '%s': %v", name, err)
			continue
		}
	}

	return nil
}

// initializeDefaultClient creates the default vault client
func (m *Manager) initializeDefaultClient() error {
	client, err := m.createClient("default", m.config)
	if err != nil {
		return err
	}

	m.mutex.Lock()
	m.clients["default"] = client
	m.mutex.Unlock()

	// Add to connection manager (this won't fail even if connection is down)
	if err := m.connectionMgr.AddConnection("default", client); err != nil {
		m.logger.Warnf("Failed to add connection to manager: %v", err)
		// Continue anyway - we'll show the connection status in the UI
	}

	// Initialize secrets manager for the default client
	m.secretsMgr = NewSecretsManager(client, m.logger)

	return nil
}

// initializeProfileClient creates a client from a vault profile
func (m *Manager) initializeProfileClient(name string, profile *config.VaultProfile) error {
	// Convert profile to VaultConfig
	vaultConfig := &config.VaultConfig{
		Address:    profile.Address,
		AuthMethod: profile.AuthMethod,
		Token:      profile.Token,
		Namespace:  profile.Namespace,
	}

	client, err := m.createClient(name, vaultConfig)
	if err != nil {
		return err
	}

	// Authenticate the client using the profile
	if err := m.authMgr.Authenticate(client.apiClient, profile); err != nil {
		m.logger.Warnf("Failed to authenticate client for '%s': %v", name, err)
		// Continue anyway - we'll show the connection status in the UI
	}

	m.mutex.Lock()
	m.clients[name] = client
	m.mutex.Unlock()

	// Add to connection manager
	if err := m.connectionMgr.AddConnection(name, client); err != nil {
		m.logger.Warnf("Failed to add connection to manager for '%s': %v", name, err)
	}

	return nil
}

// createClient creates a new vault client
func (m *Manager) createClient(name string, cfg *config.VaultConfig) (*Client, error) {
	// Create Vault API client
	apiConfig := api.DefaultConfig()
	apiConfig.Address = cfg.Address

	apiClient, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Set namespace if provided
	if cfg.Namespace != "" {
		apiClient.SetNamespace(cfg.Namespace)
	}

	client := &Client{
		apiClient: apiClient,
		config:    cfg,
		logger:    m.logger,
	}

	// Test the connection (but don't fail if it's not available)
	if err := client.TestConnection(); err != nil {
		m.logger.Warnf("Failed to connect to vault '%s': %v", name, err)
		// Continue anyway - we'll show the connection status in the UI
	}

	return client, nil
}

// GetActiveClient returns the currently active vault client
func (m *Manager) GetActiveClient() (*Client, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	client, exists := m.clients[m.activeVault]
	if !exists {
		return nil, fmt.Errorf("active vault client '%s' not found", m.activeVault)
	}

	return client, nil
}

// GetClient returns a specific vault client by name
func (m *Manager) GetClient(name string) (*Client, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	client, exists := m.clients[name]
	if !exists {
		return nil, fmt.Errorf("vault client '%s' not found", name)
	}

	return client, nil
}

// SwitchVault switches to a different vault
func (m *Manager) SwitchVault(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	client, exists := m.clients[name]
	if !exists {
		return fmt.Errorf("vault '%s' not found", name)
	}

	m.activeVault = name

	// Update secrets manager for the new active vault
	m.secretsMgr = NewSecretsManager(client, m.logger)

	m.logger.Infof("Switched to vault: %s", name)
	return nil
}

// AddVault adds a new vault connection
func (m *Manager) AddVault(name string, cfg *config.VaultConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	client, err := m.createClient(name, cfg)
	if err != nil {
		return fmt.Errorf("failed to add vault '%s': %w", name, err)
	}

	m.clients[name] = client
	m.logger.Infof("Added vault: %s", name)
	return nil
}

// ListVaults returns a list of available vault connections
func (m *Manager) ListVaults() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	vaults := make([]string, 0, len(m.clients))
	for name := range m.clients {
		vaults = append(vaults, name)
	}

	return vaults
}

// GetActiveVault returns the name of the currently active vault
func (m *Manager) GetActiveVault() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.activeVault
}

// GetConnectionManager returns the connection manager
func (m *Manager) GetConnectionManager() *ConnectionManager {
	return m.connectionMgr
}

// GetSecretsManager returns the secrets manager for the active vault
func (m *Manager) GetSecretsManager() (*SecretsManager, error) {
	client, err := m.GetActiveClient()
	if err != nil {
		return nil, err
	}
	return NewSecretsManager(client, m.logger), nil
}

// RefreshConnections refreshes all vault connections
func (m *Manager) RefreshConnections() {
	m.connectionMgr.RefreshAllConnections()
}

// GetConnectionStatus returns the status of a specific connection
func (m *Manager) GetConnectionStatus(name string) (*ConnectionStatus, error) {
	return m.connectionMgr.GetConnectionStatus(name)
}

// GetHealthyConnections returns a list of healthy connections
func (m *Manager) GetHealthyConnections() []string {
	return m.connectionMgr.GetHealthyConnections()
}

// AddVaultFromProfile adds a new vault from a profile
func (m *Manager) AddVaultFromProfile(name string, profile *config.VaultProfile) error {
	// Validate the profile
	if err := m.profilesMgr.ValidateProfile(profile); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	// Save the profile
	m.profilesMgr.SetProfile(name, profile)
	if err := m.profilesMgr.SaveProfiles(); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	// Initialize the client
	if err := m.initializeProfileClient(name, profile); err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	m.logger.Infof("Added vault from profile: %s", name)
	return nil
}

// RemoveVault removes a vault connection
func (m *Manager) RemoveVault(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if name == m.activeVault {
		return fmt.Errorf("cannot remove active vault '%s'", name)
	}

	if _, exists := m.clients[name]; !exists {
		return fmt.Errorf("vault '%s' not found", name)
	}

	// Remove from connection manager
	m.connectionMgr.RemoveConnection(name)

	// Remove from clients map
	delete(m.clients, name)

	// Remove from profiles
	if err := m.profilesMgr.DeleteProfile(name); err != nil {
		m.logger.Warnf("Failed to delete profile '%s': %v", name, err)
	}

	// Save profiles
	if err := m.profilesMgr.SaveProfiles(); err != nil {
		m.logger.Warnf("Failed to save profiles after removing '%s': %v", name, err)
	}

	m.logger.Infof("Removed vault: %s", name)
	return nil
}

// GetVaultProfiles returns all vault profiles
func (m *Manager) GetVaultProfiles() map[string]*config.VaultProfile {
	return m.profilesMgr.GetAllProfiles()
}

// GetVaultProfile returns a specific vault profile
func (m *Manager) GetVaultProfile(name string) (*config.VaultProfile, error) {
	return m.profilesMgr.GetProfile(name)
}

// UpdateVaultProfile updates an existing vault profile
func (m *Manager) UpdateVaultProfile(name string, profile *config.VaultProfile) error {
	// Validate the profile
	if err := m.profilesMgr.ValidateProfile(profile); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	// Update the profile
	m.profilesMgr.SetProfile(name, profile)
	if err := m.profilesMgr.SaveProfiles(); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	// Reinitialize the client if it exists
	if _, exists := m.clients[name]; exists {
		if err := m.initializeProfileClient(name, profile); err != nil {
			m.logger.Warnf("Failed to reinitialize client for profile '%s': %v", name, err)
		}
	}

	m.logger.Infof("Updated vault profile: %s", name)
	return nil
}

// GetVaultStatus returns detailed status information for all vaults
func (m *Manager) GetVaultStatus() map[string]*VaultStatus {
	status := make(map[string]*VaultStatus)

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name, client := range m.clients {
		connStatus, err := m.connectionMgr.GetConnectionStatus(name)
		if err != nil {
			connStatus = &ConnectionStatus{
				Connected: false,
				Error:     err.Error(),
			}
		}

		vaultStatus := &VaultStatus{
			Name:       name,
			Address:    client.config.Address,
			Namespace:  client.config.Namespace,
			AuthMethod: client.config.AuthMethod,
			Connected:  connStatus.Connected,
			Sealed:     connStatus.Sealed,
			Error:      connStatus.Error,
			Active:     name == m.activeVault,
		}

		status[name] = vaultStatus
	}

	return status
}

// VaultStatus represents the status of a vault connection
type VaultStatus struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	Namespace  string `json:"namespace"`
	AuthMethod string `json:"auth_method"`
	Connected  bool   `json:"connected"`
	Sealed     bool   `json:"sealed"`
	Error      string `json:"error,omitempty"`
	Active     bool   `json:"active"`
}
