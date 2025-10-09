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
	config              *config.Config
	vaultMgr            *vault.Manager
	uiApp               *tview.Application
	layout              *Layout
	logger              *logrus.Logger
	currentRoot         tview.Primitive // Track current screen for help dialog
	hasActiveConnection bool            // Track if user has selected a profile and has an active connection
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

	a.showVaultProfiles()

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
		case tcell.KeyTab:
			// Switch vault (alternative to Ctrl+v)
			a.switchVault()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h':
				// Show help (alternative to F1)
				a.showHelp()
				return nil
			case 'q':
				// Exit application (alternative to Ctrl+C)
				a.Stop()
				return nil
			case 'r':
				// Refresh (alternative to F5)
				a.refresh()
				return nil
			}
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
  Tab     Switch vault profiles (shows profiles table)
  Ctrl+v  Switch vault profiles (shows profiles table)
  Ctrl+n  Add new vault
  Ctrl+r  Refresh vault connection

Global:
  F1      Show help
  F5      Refresh
  Ctrl+C  Exit application

Press any key to close this help.`

	// Remember the current screen to return to it after help is closed
	previousRoot := a.currentRoot

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// Return to the screen that was active before help was opened
			if previousRoot != nil {
				a.uiApp.SetRoot(previousRoot, true)
				a.currentRoot = previousRoot
			} else {
				a.uiApp.SetRoot(a.layout.GetRoot(), true)
				a.currentRoot = a.layout.GetRoot()
			}
		})

	a.uiApp.SetRoot(modal, false)
}

// refresh refreshes the current view
func (a *App) refresh() {
	a.logger.Info("Refreshing UI")
	a.layout.Refresh()
}

// switchVault shows the vault profiles screen for switching vaults
func (a *App) switchVault() {
	a.showVaultProfiles()
}

// showVaultProfiles shows the vault profiles screen
func (a *App) showVaultProfiles() {
	// Create vault profiles panel
	profilesPanel := NewVaultProfilesPanel(a.config, a.vaultMgr, a.uiApp, a.logger)
	if err := profilesPanel.Initialize(); err != nil {
		a.logger.Errorf("Failed to initialize vault profiles panel: %v", err)
		a.showError("Failed to initialize vault profiles")
		return
	}

	// Set success callback to switch to main layout
	profilesPanel.SetSuccessCallback(func() {
		profilesPanel.StopRefresher()
		// re-initialize layout to reflect any changes in vault connections
		if err := a.layout.Initialize(); err != nil {
			a.showError(fmt.Sprintf("Failed to initialize layout: %v", err))
			return
		}
		// Mark that we now have an active connection
		a.hasActiveConnection = true
		a.currentRoot = a.layout.GetRoot()
		a.uiApp.SetRoot(a.currentRoot, true)
		// Automatically load the secrets tree after profile selection
		a.layout.Refresh()
	})

	// Set error callback to show error modal
	profilesPanel.SetErrorCallback(func(errorMsg string) {
		a.showAuthError(errorMsg)
	})

	// Create a layout with the profiles panel and a message
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Add welcome message
	var statusText, escHelpText string
	if a.hasActiveConnection {
		statusText = "Select a vault profile to switch to a different vault."
		escHelpText = "• Press Esc to go back to secrets"
	} else {
		statusText = "Please select a vault profile below to connect and begin."
		escHelpText = "• Press Enter to connect to the selected profile"
	}

	welcomeText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf(`[yellow]Welcome to VUI - Vault UI[white]

[yellow]Connection Status:[white]
%s

[yellow]Navigation:[white]
• Use arrow keys to navigate profiles
%s
• Press 'r' to refresh connection status
• Press 'n' to add a new vault profile
• Press F1 for help
• Press 'q' or Ctrl+C to exit

[yellow]Available Vault Profiles:[white]`, statusText, escHelpText))

	mainLayout.AddItem(welcomeText, 0, 1, false)
	mainLayout.AddItem(profilesPanel.GetPrimitive(), 0, 2, true)

	// Set up keyboard shortcuts for the profiles screen
	mainLayout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			// Only allow going back if user has already selected a profile (not at initial startup)
			if a.hasActiveConnection {
				profilesPanel.StopRefresher()
				a.currentRoot = a.layout.GetRoot()
				a.uiApp.SetRoot(a.currentRoot, true)
				return nil
			}
			// At initial startup, Esc does nothing - user must select a profile
			return nil
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

	a.currentRoot = mainLayout
	a.uiApp.SetRoot(mainLayout, true)
}

// showError displays an error message
func (a *App) showError(message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Error: %s", message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// Go back to vault profiles if no connected vaults
			connectedVaults := a.vaultMgr.GetConnectedConnections()
			if len(connectedVaults) == 0 {
				a.showVaultProfiles()
			} else {
				a.currentRoot = a.layout.GetRoot()
				a.uiApp.SetRoot(a.currentRoot, true)
			}
		})

	a.uiApp.SetRoot(modal, false)
}

// showAuthError displays an authentication error message and returns to vault profiles
func (a *App) showAuthError(message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[red]Authentication Failed[white]\n\n%s\n\nPlease check your credentials and vault configuration.", message)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// Always return to vault profiles screen after auth error
			if a.currentRoot != nil {
				a.uiApp.SetRoot(a.currentRoot, true)
			} else {
				a.showVaultProfiles()
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
