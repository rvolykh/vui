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
	config        *config.Config
	clients       map[string]*Client
	activeVault   string
	connectionMgr *ConnectionManager
	secretsMgr    *SecretsManager
	authMgr       *AuthManager
	mutex         sync.RWMutex
	logger        *logrus.Logger
}

// Client wraps the Vault API client with additional functionality
type Client struct {
	apiClient *api.Client
	profile   *config.VaultProfile
	logger    *logrus.Logger
}

// NewManager creates a new vault manager
func NewManager(cfg *config.Config, logger *logrus.Logger) (*Manager, error) {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	manager := &Manager{
		config:        cfg,
		clients:       make(map[string]*Client),
		activeVault:   cfg.App.DefaultVault,
		connectionMgr: NewConnectionManager(logger),
		authMgr:       NewAuthManager(logger),
		logger:        logger,
	}

	// Initialize all configured vault clients
	if err := manager.initializeAllClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize vault clients: %w", err)
	}
	// Test connections asynchronously
	manager.testAllConnectionsAsync()

	return manager, nil
}

// initializeAllClients initializes all configured vault clients
func (m *Manager) initializeAllClients() error {
	for name, profile := range m.config.Vaults {
		p := profile
		if err := m.initializeProfileClient(name, &p); err != nil {
			m.logger.Warnf("Failed to initialize client for profile '%s': %v", name, err)
			continue
		}
	}
	// Initialize secrets manager for the active client
	if m.activeVault != "" {
		if client, exists := m.clients[m.activeVault]; exists {
			m.secretsMgr = NewSecretsManager(client, m.logger)
		} else {
			m.logger.Warnf("Active vault '%s' not found in profiles, no secrets manager initialized", m.activeVault)
		}
	} else if len(m.clients) > 0 {
		// if no active vault is set, use the first one
		for name, client := range m.clients {
			m.activeVault = name
			m.secretsMgr = NewSecretsManager(client, m.logger)
			m.logger.Infof("No default vault set, using '%s' as active vault", name)
			break
		}
	}

	return nil
}

// testAllConnectionsAsync tests all vault connections asynchronously
func (m *Manager) testAllConnectionsAsync() {
	for name := range m.clients {
		m.connectionMgr.TestConnectionAsync(name)
	}
}

// initializeProfileClient creates a client from a vault profile
func (m *Manager) initializeProfileClient(name string, profile *config.VaultProfile) error {
	client, err := m.createClient(name, profile)
	if err != nil {
		return err
	}

	// Don't authenticate during initialization - authentication will happen when user selects a profile
	// This prevents unnecessary connection attempts at startup for all configured vaults

	m.mutex.Lock()
	m.clients[name] = client
	m.mutex.Unlock()

	// Add to connection manager
	m.connectionMgr.AddConnection(name, client)

	return nil
}

// createClient creates a new vault client
func (m *Manager) createClient(name string, profile *config.VaultProfile) (*Client, error) {
	// Create Vault API client
	apiConfig := api.DefaultConfig()
	apiConfig.Address = profile.Address

	apiClient, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Set namespace if provided
	if profile.Namespace != "" {
		apiClient.SetNamespace(profile.Namespace)
	}

	client := &Client{
		apiClient: apiClient,
		profile:   profile,
		logger:    m.logger,
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

	// Get the profile for authentication
	profile, ok := m.config.Vaults[name]
	if !ok {
		return fmt.Errorf("profile for vault '%s' not found", name)
	}

	// Authenticate the client when switching (if not already authenticated)
	if err := m.authMgr.Authenticate(client.apiClient, &profile); err != nil {
		m.logger.Errorf("Failed to authenticate to vault '%s': %v", name, err)
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Verify that authentication actually succeeded
	if err := m.authMgr.VerifyAuthentication(client.apiClient); err != nil {
		m.logger.Errorf("Authentication verification failed for vault '%s': %v", name, err)
		return fmt.Errorf("authentication verification failed: %w", err)
	}

	m.activeVault = name

	// Update secrets manager for the new active vault
	m.secretsMgr = NewSecretsManager(client, m.logger)

	m.logger.Infof("Switched to vault: %s", name)
	return nil
}

// AddVault adds a new vault connection
func (m *Manager) AddVault(name string, profile *config.VaultProfile) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.config.Vaults[name]; exists {
		return fmt.Errorf("vault profile '%s' already exists", name)
	}

	client, err := m.createClient(name, profile)
	if err != nil {
		return fmt.Errorf("failed to add vault '%s': %w", name, err)
	}

	m.config.Vaults[name] = *profile
	m.clients[name] = client

	if err := m.config.Save(); err != nil {
		// rollback
		delete(m.config.Vaults, name)
		delete(m.clients, name)
		return fmt.Errorf("failed to save config: %w", err)
	}
	m.connectionMgr.AddConnection(name, client)
	m.connectionMgr.TestConnectionAsync(name)

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

// GetConnectedConnections returns a list of all connected vaults (regardless of sealed/initialized status)
func (m *Manager) GetConnectedConnections() []string {
	return m.connectionMgr.GetConnectedConnections()
}

// AddVaultFromProfile adds a new vault from a profile
func (m *Manager) AddVaultFromProfile(name string, profile *config.VaultProfile) error {
	// Validate the profile
	if err := m.authMgr.ValidateAuthConfig(profile); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}
	m.config.Vaults[name] = *profile
	if err := m.config.Save(); err != nil {
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
	delete(m.config.Vaults, name)

	// Save profiles
	if err := m.config.Save(); err != nil {
		m.logger.Warnf("Failed to save config after removing '%s': %v", name, err)
	}

	m.logger.Infof("Removed vault: %s", name)
	return nil
}

// GetVaultProfiles returns all vault profiles
func (m *Manager) GetVaultProfiles() map[string]config.VaultProfile {
	return m.config.Vaults
}

// GetVaultProfile returns a specific vault profile
func (m *Manager) GetVaultProfile(name string) (*config.VaultProfile, error) {
	profile, exists := m.config.Vaults[name]
	if !exists {
		return nil, fmt.Errorf("vault profile '%s' not found", name)
	}
	return &profile, nil
}

// UpdateVaultProfile updates an existing vault profile
func (m *Manager) UpdateVaultProfile(name string, profile *config.VaultProfile) error {
	// Validate the profile
	if err := m.authMgr.ValidateAuthConfig(profile); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	// Update the profile
	m.config.Vaults[name] = *profile
	if err := m.config.Save(); err != nil {
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
			Address:    client.profile.Address,
			Namespace:  client.profile.Namespace,
			AuthMethod: client.profile.AuthMethod,
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
