package panels

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/backend"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
)

// SecretsStatus represents the status bar
type SecretsStatus struct {
	config     *config.Config
	interactor backend.Interactor
	textView   *tview.TextView
	logger     *logrus.Logger
}

// NewSecretsStatus creates a new status bar
func NewSecretsStatus(config *config.Config, interactor backend.Interactor, logger *logrus.Logger) *SecretsStatus {
	return &SecretsStatus{
		config:     config,
		interactor: interactor,
		logger:     logger,
	}
}

// Initialize initializes the status bar
func (sb *SecretsStatus) Initialize() error {
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
func (sb *SecretsStatus) UpdateConnectionStatus() {
	// Get active vault
	activeVault := sb.interactor.Profiles().GetCurrentProfile()

	// Get connection status
	status, err := sb.interactor.Profiles().GetConnectionStatus(activeVault)
	if err != nil {
		sb.updateStatus(fmt.Sprintf("[red]Error: %v[white]", err))
		return
	}

	// Format status
	var statusText string
	if status.Status == models.StatusConnected {
		if status.Status == models.StatusSealed {
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
func (sb *SecretsStatus) UpdateSelection(path string, isSecret bool) {
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
func (sb *SecretsStatus) updateStatus(text string) {
	sb.textView.SetText(text)
}

// GetPrimitive returns the underlying tview primitive
func (sb *SecretsStatus) GetPrimitive() tview.Primitive {
	return sb.textView
}
