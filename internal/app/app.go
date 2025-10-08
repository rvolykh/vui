package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/ui"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// App represents the main application
type App struct {
	config *config.Config
	vault  *vault.Manager
	logger *logrus.Logger
}

// New creates a new application instance
func New() (*App, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize vault manager
	vaultManager, err := vault.NewManager(&cfg.Vault)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize vault manager: %w", err)
	}

	return &App{
		config: cfg,
		vault:  vaultManager,
		logger: logger,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	a.logger.Info("Starting VUI application")

	// Create UI application
	uiApp := ui.NewApp(a.config, a.vault, a.logger)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		a.logger.Info("Shutting down application")
		uiApp.Stop()
	}()

	// Run the UI application
	return uiApp.Run()
}

// GetConfig returns the application configuration
func (a *App) GetConfig() *config.Config {
	return a.config
}

// GetVaultManager returns the vault manager
func (a *App) GetVaultManager() *vault.Manager {
	return a.vault
}

// GetLogger returns the application logger
func (a *App) GetLogger() *logrus.Logger {
	return a.logger
}
