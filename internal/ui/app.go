package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// App represents the UI application
type App struct {
	config   *config.Config
	vaultMgr *vault.Manager
	uiApp    *tview.Application
	layout   *Layout
	logger   *logrus.Logger
}

// NewApp creates a new UI application
func NewApp(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *App {
	uiApp := tview.NewApplication()

	// Create the main layout
	layout := NewLayout(config, vaultMgr, logger)
	layout.SetApplication(uiApp)

	return &App{
		config:   config,
		vaultMgr: vaultMgr,
		uiApp:    uiApp,
		layout:   layout,
		logger:   logger,
	}
}

// Run starts the UI application
func (a *App) Run() error {
	a.logger.Info("Starting VUI terminal interface")

	// Set up keyboard shortcuts
	a.setupKeyboardShortcuts()

	// Initialize the layout
	if err := a.layout.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize layout: %w", err)
	}

	// Check if we have any healthy connections
	healthyConnections := a.vaultMgr.GetHealthyConnections()
	if len(healthyConnections) == 0 {
		// No healthy connections, show vault profiles
		a.showVaultProfiles()
	} else {
		// We have healthy connections, show the main layout
		a.uiApp.SetRoot(a.layout.GetRoot(), true)
	}

	// Start the application
	return a.uiApp.Run()
}

// Stop stops the UI application
func (a *App) Stop() {
	a.uiApp.Stop()
}

// setupKeyboardShortcuts sets up global keyboard shortcuts
func (a *App) setupKeyboardShortcuts() {
	a.uiApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			// Exit application
			a.Stop()
			return nil
		case tcell.KeyF1:
			// Show help
			a.showHelp()
			return nil
		case tcell.KeyF5:
			// Refresh
			a.refresh()
			return nil
		case tcell.KeyCtrlV:
			// Switch vault
			a.switchVault()
			return nil
		}

		// Let other components handle the event
		return event
	})
}

// showHelp displays the help dialog
func (a *App) showHelp() {
	helpText := `VUI - Vault UI Help

Navigation:
  ↑/↓     Navigate tree items
  ←/→     Collapse/expand tree nodes
  Enter   Select item or enter directory
  Esc     Go back or cancel
  Tab     Switch between panels

Actions:
  c       Create new secret
  e       Edit selected secret
  Ctrl+d  Delete selected secret
  r       Refresh current view
  s       Search secrets
  h       Show this help
  q       Quit application

Secret Panel:
  c       Copy entire secret to clipboard
  v       Copy secret value to clipboard
  e       Edit selected secret

Vault Management:
  Ctrl+v  Switch vault
  Ctrl+n  Add new vault
  Ctrl+r  Refresh vault connection

Global:
  F1      Show help
  F5      Refresh
  Ctrl+C  Exit application

Press any key to close this help.`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.uiApp.SetRoot(a.layout.GetRoot(), true)
		})

	a.uiApp.SetRoot(modal, false)
}

// refresh refreshes the current view
func (a *App) refresh() {
	a.logger.Info("Refreshing UI")
	a.layout.Refresh()
}

// switchVault shows the vault switching dialog
func (a *App) switchVault() {
	vaults := a.vaultMgr.ListVaults()

	if len(vaults) == 0 {
		a.showError("No vaults available")
		return
	}

	// Create a list of vaults
	list := tview.NewList()
	for _, vault := range vaults {
		vault := vault // Capture loop variable
		list.AddItem(vault, "", 0, func() {
			if err := a.vaultMgr.SwitchVault(vault); err != nil {
				a.showError(fmt.Sprintf("Failed to switch to vault '%s': %v", vault, err))
				return
			}

			// Refresh the layout with new vault
			if err := a.layout.Initialize(); err != nil {
				a.showError(fmt.Sprintf("Failed to initialize layout: %v", err))
				return
			}

			a.uiApp.SetRoot(a.layout.GetRoot(), true)
		})
	}

	list.SetTitle("Select Vault").
		SetBorder(true)

	// Set up the list to handle selection
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		// The selection is handled in the AddItem callback above
	})

	list.SetDoneFunc(func() {
		a.uiApp.SetRoot(a.layout.GetRoot(), true)
	})

	a.uiApp.SetRoot(list, true)
}

// showVaultProfiles shows the vault profiles screen
func (a *App) showVaultProfiles() {
	// Create vault profiles panel
	profilesPanel := NewVaultProfilesPanel(a.config, a.vaultMgr, a.logger)
	if err := profilesPanel.Initialize(); err != nil {
		a.logger.Errorf("Failed to initialize vault profiles panel: %v", err)
		a.showError("Failed to initialize vault profiles")
		return
	}

	// Set success callback to switch to main layout
	profilesPanel.SetSuccessCallback(func() {
		a.uiApp.SetRoot(a.layout.GetRoot(), true)
	})

	// Create a layout with the profiles panel and a message
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Add welcome message
	welcomeText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[yellow]Welcome to VUI - Vault UI[white]

[yellow]Connection Status:[white]
No vault servers are currently connected. Please select a vault profile below to connect.

[yellow]Navigation:[white]
• Use arrow keys to navigate
• Press Enter to connect to a vault
• Press 'r' to refresh connection status
• Press 'n' to add a new vault profile
• Press F1 for help
• Press Ctrl+C to exit

[yellow]Available Vault Profiles:[white]`)

	mainLayout.AddItem(welcomeText, 0, 1, false)
	mainLayout.AddItem(profilesPanel.GetPrimitive(), 0, 2, true)

	// Set up keyboard shortcuts for the profiles screen
	mainLayout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			// Show help
			a.showHelp()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				// Quit
				a.Stop()
				return nil
			}
		}

		// Let the profiles panel handle other events
		return event
	})

	a.uiApp.SetRoot(mainLayout, true)
}

// showError displays an error message
func (a *App) showError(message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Error: %s", message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// Go back to vault profiles if no healthy connections
			healthyConnections := a.vaultMgr.GetHealthyConnections()
			if len(healthyConnections) == 0 {
				a.showVaultProfiles()
			} else {
				a.uiApp.SetRoot(a.layout.GetRoot(), true)
			}
		})

	a.uiApp.SetRoot(modal, false)
}

// GetUIApp returns the underlying tview application
func (a *App) GetUIApp() *tview.Application {
	return a.uiApp
}

// GetLayout returns the main layout
func (a *App) GetLayout() *Layout {
	return a.layout
}
