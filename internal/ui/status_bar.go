package ui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// StatusBar represents the status bar
type StatusBar struct {
	config   *config.Config
	vaultMgr *vault.Manager
	textView *tview.TextView
	logger   *logrus.Logger
}

// NewStatusBar creates a new status bar
func NewStatusBar(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *StatusBar {
	return &StatusBar{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
	}
}

// Initialize initializes the status bar
func (sb *StatusBar) Initialize() error {
	sb.textView = tview.NewTextView()

	// Set up the text view appearance
	sb.textView.SetBorder(false)

	// Enable dynamic colors
	sb.textView.SetDynamicColors(true)

	// Set initial status
	sb.UpdateConnectionStatus()

	return nil
}

// UpdateConnectionStatus updates the connection status
func (sb *StatusBar) UpdateConnectionStatus() {
	// Get active vault
	activeVault := sb.vaultMgr.GetActiveVault()

	// Get connection status
	status, err := sb.vaultMgr.GetConnectionStatus(activeVault)
	if err != nil {
		sb.updateStatus(fmt.Sprintf("[red]Error: %v[white]", err))
		return
	}

	// Format status
	var statusText string
	if status.Connected {
		if status.Sealed {
			statusText = fmt.Sprintf("[yellow]Vault: %s (Sealed)[white]", activeVault)
		} else {
			statusText = fmt.Sprintf("[green]Vault: %s (Connected)[white]", activeVault)
		}
	} else {
		statusText = fmt.Sprintf("[red]Vault: %s (Disconnected)[white]", activeVault)
	}

	// Add version info if available
	if status.Version != "" {
		statusText += fmt.Sprintf(" | Version: %s", status.Version)
	}

	// Add cluster info if available
	if status.ClusterID != "" {
		statusText += fmt.Sprintf(" | Cluster: %s", status.ClusterID)
	}

	// Add timestamp
	statusText += fmt.Sprintf(" | %s", time.Now().Format("15:04:05"))

	sb.updateStatus(statusText)
}

// UpdateSelection updates the selection status
func (sb *StatusBar) UpdateSelection(path string, isSecret bool) {
	var itemType string
	if isSecret {
		itemType = "Secret"
	} else {
		itemType = "Directory"
	}

	selectionText := fmt.Sprintf("[blue]Selected: %s (%s)[white]", path, itemType)
	sb.updateStatus(selectionText)
}

// updateStatus updates the status text
func (sb *StatusBar) updateStatus(text string) {
	sb.textView.SetText(text)
}

// GetPrimitive returns the underlying tview primitive
func (sb *StatusBar) GetPrimitive() tview.Primitive {
	return sb.textView
}
