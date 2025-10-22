package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/ui/common"
	"github.com/rvolykh/vui/internal/ui/panels"
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
	dialogSvc           *common.DialogService
	currentRoot         tview.Primitive // Track current screen for help dialog
	hasActiveConnection bool            // Track if user has selected a profile and has an active connection
	onProfilesScreen    bool            // Track if we're currently on the profiles screen
}

// NewApp creates a new UI application
func NewApp(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *App {
	// Initialize theme
	common.InitializeTheme(config.UI.Theme)

	uiApp := tview.NewApplication()

	// Create the main layout
	layout := NewLayout(config, vaultMgr, logger)
	layout.SetApplication(uiApp)

	// Create dialog service
	dialogSvc := common.NewDialogService(uiApp, nil) // Root will be set later

	return &App{
		config:    config,
		vaultMgr:  vaultMgr,
		uiApp:     uiApp,
		layout:    layout,
		logger:    logger,
		dialogSvc: dialogSvc,
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
		// Check if a modal/form is currently active
		hasModal := a.layout != nil && a.layout.HasActiveModal()

		switch event.Key() {
		case tcell.KeyCtrlC:
			// Exit application - always allow
			a.Stop()
			return nil
		case tcell.KeyF1:
			// Show help - disable in forms
			if !hasModal {
				a.showHelp()
				return nil
			}
			return event
		case tcell.KeyF5:
			// Refresh - context aware
			// Disable in forms, if on profiles screen let the profiles panel handle it
			if hasModal {
				return event
			}
			if !a.onProfilesScreen {
				a.refresh()
				return nil
			}
			// Let the profiles panel handle F5
			return event
		case tcell.KeyCtrlV:
			// Switch vault - disable in forms
			if !hasModal {
				a.switchVault()
				return nil
			}
			return event
		case tcell.KeyTab:
			// Switch vault (alternative to Ctrl+v)
			// But only if we're not in a form/modal - let forms handle Tab for field navigation
			if !a.onProfilesScreen && !hasModal {
				a.switchVault()
				return nil
			}
			// Let the form/modal handle Tab
			return event
		case tcell.KeyRune:
			// Disable all single-letter hotkeys when a form is active
			if hasModal {
				return event
			}

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
				// Refresh (alternative to F5) - context aware
				// If on profiles screen, let the profiles panel handle it
				if !a.onProfilesScreen {
					a.refresh()
					return nil
				}
				// Let the profiles panel handle 'r'
				return event
			}
		}

		// Let other components handle the event
		return event
	})
}

// showHelp displays the help dialog
func (a *App) showHelp() {
	helpText := `VUI - Vault UI Help

[b]Navigation:[white]
  ↑/↓      Navigate tree items                                          
  ←/→      Collapse/expand tree nodes                                   
  Enter    Select item or enter directory                               
  Esc      Go back or cancel                                            
  Tab      Navigate form fields (in forms) / Switch vault (in main view)

[b]Secret Panel:[white]
  c        Create new secret             
  e        Edit selected secret          
  Ctrl+d   Delete selected secret        
  d        Unmask/mask secret value      
  v        Copy secret value to clipboard

[b]Vault Management:[white]
  Tab      Switch vault profiles (shows profiles table)         
  Esc      Go back to secrets (if previously selected a profile)

[b]Global:[white]
  h / F1       Show help       
  r / F5       Refresh         
  q / Ctrl+C   Exit application
`

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
		}).
		SetButtonBackgroundColor(tcell.ColorDarkGray).
		SetButtonTextColor(tcell.ColorWhite).
		SetButtonActivatedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorBlack))

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
	// Mark that we're on the profiles screen
	a.onProfilesScreen = true

	// Create vault profiles panel
	profilesPanel := panels.NewProfilesTable(a.config, a.vaultMgr, a.uiApp, a.logger)
	if err := profilesPanel.Initialize(); err != nil {
		a.logger.Errorf("Failed to initialize vault profiles panel: %v", err)
		a.showError("Failed to initialize vault profiles")
		return
	}

	// Set success callback to switch to main layout
	profilesPanel.SetSuccessCallback(func() {
		profilesPanel.StopRefresher()
		// Mark that we're no longer on the profiles screen
		a.onProfilesScreen = false
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
	welcomeText := a.buildWelcomeText()

	mainLayout.AddItem(welcomeText, 0, 1, false)
	mainLayout.AddItem(profilesPanel.GetPrimitive(), 0, 2, true)

	// Set up keyboard shortcuts for the profiles screen
	mainLayout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			// Only allow going back if user has already selected a profile (not at initial startup)
			if a.hasActiveConnection {
				profilesPanel.StopRefresher()
				// Mark that we're no longer on the profiles screen
				a.onProfilesScreen = false
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
	modal := common.ErrorModal(message, func() {
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
	errorMsg := fmt.Sprintf("Authentication Failed\n\n%s\n\nPlease check your credentials and vault configuration.", message)
	modal := common.ErrorModal(errorMsg, func() {
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

// buildWelcomeText creates a formatted welcome panel with connection status and navigation info
func (a *App) buildWelcomeText() *tview.TextView {
	welcomeScreen := panels.NewProfilesTitle(a.vaultMgr, a.hasActiveConnection)
	return welcomeScreen.Build()
}
