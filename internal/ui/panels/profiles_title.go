package panels

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/backend"
	"github.com/rvolykh/vui/internal/ui/common"
)

// ProfilesTitle represents the welcome/connection screen
type ProfilesTitle struct {
	interactor    backend.Interactor
	hasActiveConn bool
}

// NewProfilesTitle creates a new welcome screen
func NewProfilesTitle(interactor backend.Interactor, hasActiveConnection bool) *ProfilesTitle {
	return &ProfilesTitle{
		interactor:    interactor,
		hasActiveConn: hasActiveConnection,
	}
}

// Build creates the welcome screen text view
func (ws *ProfilesTitle) Build() *tview.TextView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false)

	// Set a draw function that rebuilds the content based on available width
	textView.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// Build the content with the actual width
		content := ws.buildContent(width)
		textView.SetText(content)
		return x, y, width, height
	})

	// Initialize with a default width
	textView.SetText(ws.buildContent(80))

	return textView
}

// buildContent generates the welcome text content for a given width
func (ws *ProfilesTitle) buildContent(width int) string {
	// Ensure minimum width
	if width < 60 {
		width = 60
	}

	// Maximum width for readability
	if width > 120 {
		width = 120
	}

	// Get connection status
	connectionStatus := ws.getConnectionStatus()

	// Define navigation keys and their descriptions
	navigationItems := ws.getNavigationItems()

	// Define configuration paths
	configPaths := []string{
		"./configs/vui.yaml",
		"$HOME/.vui/vui.yaml",
		"/etc/vui/vui.yaml",
	}

	// Calculate inner width (excluding borders)
	innerWidth := width - 4 // 2 for borders + 2 for padding

	// Build the welcome text with proper formatting
	var b string

	// Top border
	b += "┌" + common.RepeatString("─", width-2) + "┐\n"

	// Title
	b += "│ " + common.PadRight("[yellow::b]Welcome to VUI - Vault UI[white]", innerWidth) + " │\n"
	b += "│ " + common.RepeatString(" ", innerWidth) + " │\n"

	// Connection Status header
	b += "│ " + common.PadRight("[yellow]Connection Status:[white]", innerWidth) + " │\n"
	b += "│ " + common.PadRight(connectionStatus, innerWidth) + " │\n"
	b += "│ " + common.RepeatString(" ", innerWidth) + " │\n"

	// Two-column headers
	leftColWidth := innerWidth / 2
	rightColWidth := innerWidth - leftColWidth
	b += "│ " + common.PadRight("[yellow]Navigation[white]", leftColWidth) + common.PadRight("[yellow]Config Paths[white]", rightColWidth) + " │\n"

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
			leftContent = keysPart + common.RepeatString(" ", keysPadding) + nav.desc
		}

		// Right column: Configuration
		rightContent := ""
		if i < len(configPaths) {
			rightContent = fmt.Sprintf("  [green]- %s[white]", configPaths[i])
		}

		line := "│ " + common.PadRight(leftContent, leftColWidth) + common.PadRight(rightContent, rightColWidth) + " │\n"
		b += line
	}

	// Bottom border
	b += "└" + common.RepeatString("─", width-2) + "┘\n"
	b += "\n[yellow]Available Vault Profiles:[white]"

	return b
}

// getConnectionStatus returns the formatted connection status string
func (ws *ProfilesTitle) getConnectionStatus() string {
	if ws.hasActiveConn {
		activeVault := ws.interactor.Profiles().GetCurrentProfile()
		if activeVault != "" {
			return fmt.Sprintf("[green]Connected:[white] %s", activeVault)
		}
		return "[yellow]No active connection - please select a profile below[white]"
	}
	return "[yellow]No connection - please select a profile below to begin[white]"
}

// navigationItem holds a navigation key and description
type navigationItem struct {
	keys string
	desc string
}

// getNavigationItems returns the list of navigation items
func (ws *ProfilesTitle) getNavigationItems() []navigationItem {
	items := []navigationItem{
		{"↑/↓", "Navigate profiles"},
		{"Enter", "Connect to profile"},
		{"r/F5", "Refresh status"},
		{"h/F1", "Show help"},
		{"q/Ctrl+C", "Exit"},
	}

	// Add Esc option if we have an active connection
	if ws.hasActiveConn {
		items = append([]navigationItem{{"Esc", "Back to secrets"}}, items...)
	}

	return items
}
