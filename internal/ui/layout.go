package ui

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// Layout represents the main application layout
type Layout struct {
	config      *config.Config
	vaultMgr    *vault.Manager
	root        *tview.Flex
	treePanel   *TreePanel
	secretPanel *SecretPanel
	statusBar   *StatusBar
	modal       tview.Primitive
	app         *tview.Application
	logger      *logrus.Logger
}

// NewLayout creates a new layout
func NewLayout(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *Layout {
	return &Layout{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
	}
}

// SetApplication sets the tview application reference
func (l *Layout) SetApplication(app *tview.Application) {
	l.app = app
}

// Initialize initializes the layout components
func (l *Layout) Initialize() error {
	l.logger.Info("Initializing UI layout")

	// Check if we have any connected vaults (even if sealed or not fully initialized)
	connectedVaults := l.vaultMgr.GetConnectedConnections()
	if len(connectedVaults) == 0 {
		l.logger.Info("No vault connections available, initializing layout in offline mode")
		return l.initializeOfflineMode()
	}

	// Create the tree panel
	l.treePanel = NewTreePanel(l.config, l.vaultMgr, l.logger)
	if err := l.treePanel.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize tree panel: %w", err)
	}

	// Create the secret panel
	l.secretPanel = NewSecretPanel(l.config, l.vaultMgr, l.logger)
	if err := l.secretPanel.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize secret panel: %w", err)
	}

	// Create the status bar
	l.statusBar = NewStatusBar(l.config, l.vaultMgr, l.logger)
	if err := l.statusBar.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize status bar: %w", err)
	}

	// Set up the layout
	l.setupLayout()

	// Set up event handlers
	l.setupEventHandlers()

	return nil
}

// initializeOfflineMode initializes the layout in offline mode
func (l *Layout) initializeOfflineMode() error {
	// Create a simple offline layout
	l.root = tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Create offline message
	offlineText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[yellow]VUI - Vault UI (Offline Mode)[white]

[red]No vault connections available[white]

The application is running in offline mode because no vault servers are currently connected.

[yellow]To connect to a vault:[white]
1. Ensure your vault server is running
2. Check your configuration in ~/.vui/config.yaml
3. Verify network connectivity to your vault server
4. Restart the application

[yellow]Configuration:[white]
• Default vault address: ` + l.getDefaultVaultAddress() + `
• Auth method: ` + l.getDefaultVaultAuthMethod() + `

[yellow]Press Ctrl+C to exit[white]`)

	l.root.AddItem(offlineText, 0, 1, true)

	return nil
}

// getDefaultVaultAddress returns the address of the default vault
func (l *Layout) getDefaultVaultAddress() string {
	defaultVaultName := l.config.App.DefaultVault
	if defaultVaultName == "" {
		defaultVaultName = "default"
	}
	if profile, ok := l.config.Vaults[defaultVaultName]; ok {
		return profile.Address
	}
	return "not configured"
}

// getDefaultVaultAuthMethod returns the auth method of the default vault
func (l *Layout) getDefaultVaultAuthMethod() string {
	defaultVaultName := l.config.App.DefaultVault
	if defaultVaultName == "" {
		defaultVaultName = "default"
	}
	if profile, ok := l.config.Vaults[defaultVaultName]; ok {
		return profile.AuthMethod
	}
	return "not configured"
}

// setupLayout creates the main layout structure
func (l *Layout) setupLayout() {
	// Create the main horizontal layout
	l.root = tview.NewFlex().
		SetDirection(tview.FlexColumn)

	// Add the tree panel (left side)
	l.root.AddItem(l.treePanel.GetPrimitive(), 0, 1, true)

	// Add the secret panel (right side)
	l.root.AddItem(l.secretPanel.GetPrimitive(), 0, 2, false)

	// Create the main vertical layout
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(l.root, 0, 1, true).
		AddItem(l.statusBar.GetPrimitive(), 1, 0, false)

	l.root = mainLayout
}

// setupEventHandlers sets up event handlers between components
func (l *Layout) setupEventHandlers() {
	// When a tree item is selected, update the secret panel
	l.treePanel.SetSelectionHandler(func(path string, isSecret bool) {
		if isSecret {
			l.secretPanel.ShowSecret(path)
		} else {
			l.secretPanel.ShowDirectory(path)
		}
		l.statusBar.UpdateSelection(path, isSecret)
	})

	// When tree is refreshed, update status bar
	l.treePanel.SetRefreshHandler(func() {
		l.statusBar.UpdateConnectionStatus()
	})

	// Set up modal handler for tree panel
	l.treePanel.SetModalHandler(func(primitive tview.Primitive, show bool) {
		l.showModal(primitive, show)
	})
}

// GetRoot returns the root primitive
func (l *Layout) GetRoot() tview.Primitive {
	return l.root
}

// Refresh refreshes all layout components
func (l *Layout) Refresh() {
	l.logger.Info("Refreshing layout")

	// Refresh tree panel
	if l.treePanel != nil {
		l.treePanel.Refresh()
	}

	// Refresh secret panel
	if l.secretPanel != nil {
		l.secretPanel.Refresh()
	}

	// Refresh status bar
	if l.statusBar != nil {
		l.statusBar.UpdateConnectionStatus()
	}
}

// GetTreePanel returns the tree panel
func (l *Layout) GetTreePanel() *TreePanel {
	return l.treePanel
}

// GetSecretPanel returns the secret panel
func (l *Layout) GetSecretPanel() *SecretPanel {
	return l.secretPanel
}

// GetStatusBar returns the status bar
func (l *Layout) GetStatusBar() *StatusBar {
	return l.statusBar
}

// showModal shows or hides a modal dialog
func (l *Layout) showModal(primitive tview.Primitive, show bool) {
	if l.app == nil {
		l.logger.Warn("Cannot show modal: application reference not set")
		return
	}

	if show {
		l.modal = primitive
		l.app.SetRoot(primitive, true)
	} else {
		l.modal = nil
		l.app.SetRoot(l.root, true)
	}
}
