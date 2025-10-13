package ui

import (
	"fmt"
	"unicode/utf8"

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
	onProfilesScreen    bool            // Track if we're currently on the profiles screen
}

// NewApp creates a new UI application
func NewApp(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *App {
	// Customize tview theme colors before creating UI elements
	// Replace the default purple/magenta theme with a more neutral dark theme
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack        // Main background
	tview.Styles.ContrastBackgroundColor = tcell.ColorDarkSlateGray // Input fields, buttons (was purple)
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorDarkGray  // Even more contrast
	tview.Styles.BorderColor = tcell.ColorDarkCyan                  // Border color
	tview.Styles.TitleColor = tcell.ColorWhite                      // Title text
	tview.Styles.GraphicsColor = tcell.ColorDarkCyan                // Graphics elements
	tview.Styles.PrimaryTextColor = tcell.ColorWhite                // Primary text
	tview.Styles.SecondaryTextColor = tcell.ColorLightGray          // Secondary text
	tview.Styles.TertiaryTextColor = tcell.ColorGray                // Tertiary text
	tview.Styles.InverseTextColor = tcell.ColorBlack                // Inverse text
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorDarkGray   // Contrast secondary text

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

Navigation:
  ↑/↓     Navigate tree items
  ←/→     Collapse/expand tree nodes
  Enter   Select item or enter directory
  Esc     Go back or cancel
  Tab     Navigate form fields (in forms) / Switch vault (in main view)

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
	profilesPanel := NewVaultProfilesPanel(a.config, a.vaultMgr, a.uiApp, a.logger)
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
		}).
		SetButtonBackgroundColor(tcell.ColorDarkGray).
		SetButtonTextColor(tcell.ColorWhite).
		SetButtonActivatedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorBlack))

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
		}).
		SetButtonBackgroundColor(tcell.ColorDarkGray).
		SetButtonTextColor(tcell.ColorWhite).
		SetButtonActivatedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorBlack))

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
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false)

	// Set a draw function that rebuilds the content based on available width
	textView.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// Build the content with the actual width
		content := a.buildWelcomeContent(width)
		textView.SetText(content)
		return x, y, width, height
	})

	// Initialize with a default width
	textView.SetText(a.buildWelcomeContent(80))

	return textView
}

// buildWelcomeContent generates the welcome text content for a given width
func (a *App) buildWelcomeContent(width int) string {
	// Ensure minimum width
	if width < 60 {
		width = 60
	}

	// Maximum width for readability
	if width > 120 {
		width = 120
	}

	// Get connection status
	var connectionStatus string
	if a.hasActiveConnection {
		activeVault := a.vaultMgr.GetActiveVault()
		if activeVault != "" {
			connectionStatus = fmt.Sprintf("[green]Connected:[white] %s", activeVault)
		} else {
			connectionStatus = "[yellow]No active connection - please select a profile below[white]"
		}
	} else {
		connectionStatus = "[yellow]No connection - please select a profile below to begin[white]"
	}

	// Define navigation keys and their descriptions
	navigationItems := []struct {
		keys string
		desc string
	}{
		{"↑/↓", "Navigate profiles"},
		{"Enter", "Connect to profile"},
		{"r/F5", "Refresh status"},
		{"n", "Add new profile"},
		{"F1", "Show help"},
		{"q/Ctrl+C", "Exit"},
	}

	// Add Esc option if we have an active connection
	if a.hasActiveConnection {
		navigationItems = append([]struct {
			keys string
			desc string
		}{{"Esc", "Back to secrets"}}, navigationItems...)
	}

	// Define configuration paths
	configPaths := []string{
		"./configs/default.yaml",
		"$HOME/.vui/default.yaml",
		"/etc/vui/default.yaml",
	}

	// Calculate inner width (excluding borders)
	innerWidth := width - 4 // 2 for borders + 2 for padding

	// Build the welcome text with proper formatting
	var b string

	// Top border
	b += "┌" + repeatString("─", width-2) + "┐\n"

	// Title
	b += "│ " + padRight("[yellow::b]Welcome to VUI - Vault UI[white]", innerWidth) + " │\n"
	b += "│ " + repeatString(" ", innerWidth) + " │\n"

	// Connection Status header
	b += "│ " + padRight("[yellow]Connection Status:[white]", innerWidth) + " │\n"
	b += "│ " + padRight(connectionStatus, innerWidth) + " │\n"
	b += "│ " + repeatString(" ", innerWidth) + " │\n"

	// Two-column headers
	leftColWidth := innerWidth / 2
	rightColWidth := innerWidth - leftColWidth
	b += "│ " + padRight("[yellow]Navigation[white]", leftColWidth) + padRight("[yellow]Config Paths[white]", rightColWidth) + " │\n"

	// Add navigation items and config paths side by side
	maxRows := len(navigationItems)
	if len(configPaths) > maxRows {
		maxRows = len(configPaths)
	}

	for i := 0; i < maxRows; i++ {
		// Left column: Navigation
		leftContent := ""
		if i < len(navigationItems) {
			nav := navigationItems[i]
			// Build left content with explicit padding for keys (12 chars total including indentation)
			keysPart := fmt.Sprintf("  [cyan]%s[white]", nav.keys)
			// Pad keys to consistent width (use rune count for proper Unicode handling)
			keysVisible := utf8.RuneCountInString(nav.keys)
			keysPadding := 12 - keysVisible
			if keysPadding < 0 {
				keysPadding = 0
			}
			leftContent = keysPart + repeatString(" ", keysPadding) + nav.desc
		}

		// Right column: Configuration
		rightContent := ""
		if i < len(configPaths) {
			rightContent = fmt.Sprintf("  [green]- %s[white]", configPaths[i])
		}

		line := "│ " + padRight(leftContent, leftColWidth) + padRight(rightContent, rightColWidth) + " │\n"
		b += line
	}

	// Bottom border
	b += "└" + repeatString("─", width-2) + "┘\n"
	b += "\n[yellow]Available Vault Profiles:[white]"

	return b
}

// padRight pads a string to the right with spaces, accounting for tview color tags
func padRight(s string, width int) string {
	visibleLen := utf8.RuneCountInString(s) - countColorTags(s)
	if visibleLen >= width {
		return s
	}
	return s + repeatString(" ", width-visibleLen)
}

// repeatString repeats a string n times
func repeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// countColorTags counts the number of characters used by tview color tags
func countColorTags(s string) int {
	count := 0
	inTag := false
	for _, c := range s {
		if c == '[' {
			inTag = true
		}
		if inTag {
			count++
		}
		if c == ']' {
			inTag = false
		}
	}
	return count
}
