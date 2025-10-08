package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// VaultProfilesPanel displays vault profiles with connection status
type VaultProfilesPanel struct {
	config          *config.Config
	vaultMgr        *vault.Manager
	list            *tview.List
	logger          *logrus.Logger
	successCallback func() // Callback to switch to main layout
}

// NewVaultProfilesPanel creates a new vault profiles panel
func NewVaultProfilesPanel(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger) *VaultProfilesPanel {
	return &VaultProfilesPanel{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
	}
}

// SetSuccessCallback sets the callback to be called when a vault is successfully connected
func (vpp *VaultProfilesPanel) SetSuccessCallback(callback func()) {
	vpp.successCallback = callback
}

// Initialize initializes the vault profiles panel
func (vpp *VaultProfilesPanel) Initialize() error {
	vpp.list = tview.NewList()

	// Set up the list appearance
	vpp.list.SetBorder(true).
		SetTitle("Vault Profiles").
		SetTitleAlign(tview.AlignLeft)

	// Set up keyboard navigation
	vpp.setupKeyboardNavigation()

	// Load and display profiles
	return vpp.loadProfiles()
}

// setupKeyboardNavigation sets up keyboard navigation
func (vpp *VaultProfilesPanel) setupKeyboardNavigation() {
	vpp.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			// Switch to selected vault
			vpp.switchToSelectedVault()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'r':
				// Refresh profiles
				vpp.Refresh()
				return nil
			case 'n':
				// Add new vault
				vpp.addNewVault()
				return nil
			case 'd':
				// Delete selected vault (with Ctrl)
				if event.Modifiers()&tcell.ModCtrl != 0 {
					vpp.deleteSelectedVault()
					return nil
				}
			}
		}

		return event
	})
}

// loadProfiles loads and displays vault profiles
func (vpp *VaultProfilesPanel) loadProfiles() error {
	// Clear existing items
	vpp.list.Clear()

	// Get available vaults
	vaults := vpp.vaultMgr.ListVaults()

	if len(vaults) == 0 {
		vpp.list.AddItem("No vault profiles configured", "", 0, nil)
		return nil
	}

	// Add each vault profile
	for _, vaultName := range vaults {
		vaultName := vaultName // Capture loop variable

		// Get connection status
		status, err := vpp.vaultMgr.GetConnectionStatus(vaultName)
		if err != nil {
			vpp.logger.Warnf("Failed to get status for vault '%s': %v", vaultName, err)
			continue
		}

		// Format the display text
		displayText := vpp.formatVaultDisplay(vaultName, status)

		// Add to list
		vpp.list.AddItem(displayText, "", 0, func() {
			vpp.switchToVault(vaultName)
		})
	}

	return nil
}

// formatVaultDisplay formats the vault display text with status
func (vpp *VaultProfilesPanel) formatVaultDisplay(name string, status *vault.ConnectionStatus) string {
	var statusIcon string
	var statusText string

	if status.Connected {
		if status.Sealed {
			statusIcon = "🔒"
			statusText = "Sealed"
		} else {
			statusIcon = "✅"
			statusText = "Connected"
		}
	} else {
		statusIcon = "❌"
		statusText = "Disconnected"
	}

	// Build the display text
	var parts []string
	parts = append(parts, fmt.Sprintf("%s %s", statusIcon, name))
	parts = append(parts, fmt.Sprintf("Status: %s", statusText))
	parts = append(parts, fmt.Sprintf("Address: %s", status.Address))

	if status.Version != "" {
		parts = append(parts, fmt.Sprintf("Version: %s", status.Version))
	}

	if status.Error != "" {
		parts = append(parts, fmt.Sprintf("Error: %s", status.Error))
	}

	return strings.Join(parts, " | ")
}

// switchToSelectedVault switches to the currently selected vault
func (vpp *VaultProfilesPanel) switchToSelectedVault() {
	currentItem := vpp.list.GetCurrentItem()
	if currentItem < 0 {
		return
	}

	// Get the vault name from the display text
	mainText, _ := vpp.list.GetItemText(currentItem)
	vaultName := vpp.extractVaultName(mainText)

	if vaultName != "" {
		vpp.switchToVault(vaultName)
	}
}

// switchToVault switches to a specific vault
func (vpp *VaultProfilesPanel) switchToVault(vaultName string) {
	if err := vpp.vaultMgr.SwitchVault(vaultName); err != nil {
		vpp.logger.Errorf("Failed to switch to vault '%s': %v", vaultName, err)
		return
	}

	vpp.logger.Infof("Switched to vault: %s", vaultName)

	// Refresh connection status
	vpp.vaultMgr.RefreshConnections()

	// Check if the vault is now healthy
	healthyConnections := vpp.vaultMgr.GetHealthyConnections()
	if len(healthyConnections) > 0 {
		// We have a healthy connection, switch to main layout
		if vpp.successCallback != nil {
			vpp.successCallback()
		}
	} else {
		// Still no healthy connections, refresh the display
		vpp.Refresh()
	}
}

// addNewVault adds a new vault profile
func (vpp *VaultProfilesPanel) addNewVault() {
	// This would show a form to add a new vault
	// For now, just log the action
	vpp.logger.Info("Add new vault (not implemented yet)")
}

// deleteSelectedVault deletes the selected vault profile
func (vpp *VaultProfilesPanel) deleteSelectedVault() {
	currentItem := vpp.list.GetCurrentItem()
	if currentItem < 0 {
		return
	}

	// Get the vault name from the display text
	mainText, _ := vpp.list.GetItemText(currentItem)
	vaultName := vpp.extractVaultName(mainText)

	if vaultName != "" {
		vpp.logger.Infof("Delete vault: %s (not implemented yet)", vaultName)
	}
}

// extractVaultName extracts the vault name from the display text
func (vpp *VaultProfilesPanel) extractVaultName(displayText string) string {
	// The display text format is: "ICON VaultName | Status: ... | Address: ..."
	parts := strings.Split(displayText, " | ")
	if len(parts) > 0 {
		// Remove the icon and get the vault name
		firstPart := strings.TrimSpace(parts[0])
		if len(firstPart) > 2 {
			return firstPart[2:] // Skip the icon (2 characters)
		}
	}
	return ""
}

// Refresh refreshes the profiles display
func (vpp *VaultProfilesPanel) Refresh() {
	vpp.logger.Info("Refreshing vault profiles")

	// Refresh connection statuses
	vpp.vaultMgr.RefreshConnections()

	// Reload profiles
	if err := vpp.loadProfiles(); err != nil {
		vpp.logger.Errorf("Failed to refresh profiles: %v", err)
	}
}

// GetPrimitive returns the underlying tview primitive
func (vpp *VaultProfilesPanel) GetPrimitive() tview.Primitive {
	return vpp.list
}
