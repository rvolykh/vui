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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger := logrus.New()
	logLevel, err := logrus.ParseLevel(cfg.App.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logFile, err := os.OpenFile(cfg.App.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	logger.SetOutput(logFile)

	// Initialize vault manager
	vaultManager, err := vault.NewManager(cfg, logger)
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
