package backend

import (
	"fmt"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/engines"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
)

type ProfileInteractor interface {
	SwitchProfile(name string) error
	GetCurrentProfile() string

	ListConnections() []string
	RefreshConnection(name string)
	ResetConnections()
	ReloadConfiguration() error
	GetConnectionStatus(name string) (*models.ConnectionStatus, error)
}

type profileInteractor struct {
	config            *config.Config
	currentProfile    string
	secretsInteractor SecretsInteractor
	connectionMgr     *ConnectionManager
	logger            *logrus.Logger
}

func newProfileInteractor(logger *logrus.Logger, cfg *config.Config) (*profileInteractor, error) {
	interactor := &profileInteractor{
		config:        cfg,
		connectionMgr: NewConnectionManager(logger),
		logger:        logger,
	}

	// Initialize all configured vault clients
	if err := interactor.initializeConnections(); err != nil {
		return nil, fmt.Errorf("failed to initialize profiles: %w", err)
	}
	// Test connections asynchronously
	interactor.testAllConnectionsAsync()

	return interactor, nil
}

// SwitchVault switches to a different vault
func (i *profileInteractor) SwitchProfile(name string) error {
	client, err := i.connectionMgr.GetConnection(name)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Authenticate the client when switching
	if err := client.Authenticate(); err != nil {
		i.logger.Errorf("Failed to authenticate to vault '%s': %v", name, err)
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Update secrets manager for the new active vault
	i.currentProfile = name
	i.secretsInteractor = newSecretsInteractor(i.logger, name, client)

	i.logger.Infof("Switched to vault: %s", name)
	return nil
}

func (i *profileInteractor) GetCurrentProfile() string {
	return i.currentProfile
}

func (i *profileInteractor) ListConnections() []string {
	return i.connectionMgr.ListConnections()
}

func (i *profileInteractor) RefreshConnection(name string) {
	i.connectionMgr.RefreshConnectionStatus(name)
}

func (i *profileInteractor) ResetConnections() {
	i.connectionMgr.ResetConnections()
}

func (i *profileInteractor) GetConnectionStatus(name string) (*models.ConnectionStatus, error) {
	return i.connectionMgr.GetConnectionStatus(name)
}

func (i *profileInteractor) ReloadConfiguration() error {
	i.logger.Info("Reloading configuration from disk...")

	// Reload config from disk
	newConfig, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}
	i.config = newConfig

	// Remove all connections
	for _, name := range i.connectionMgr.ListConnections() {
		i.connectionMgr.RemoveConnection(name)
	}
	i.currentProfile = ""

	// Initialize all configured vault clients
	if err := i.initializeConnections(); err != nil {
		return fmt.Errorf("failed to initialize profiles: %w", err)
	}

	// Test connections asynchronously
	i.logger.Info("Re-testing all vault connections...")
	i.testAllConnectionsAsync()

	i.logger.Info("Configuration reloaded successfully")
	return nil
}

func (i *profileInteractor) initializeConnections() error {
	factory := engines.NewEnginesFactory(i.logger)

	for name, profile := range i.config.Vaults {
		p := profile

		client, err := factory.SetupEngine("vault", &p) // TODO: add support for other engines
		if err != nil {
			i.logger.Warnf("Failed to setup engine for profile '%s': %v", name, err)
			continue
		}

		i.connectionMgr.AddConnection(name, client)
	}

	return nil
}

func (i *profileInteractor) testAllConnectionsAsync() {
	for _, name := range i.connectionMgr.ListConnections() {
		i.connectionMgr.TestConnectionAsync(name)
	}
}
