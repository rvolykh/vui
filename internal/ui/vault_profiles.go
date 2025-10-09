package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

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
	app             *tview.Application
	logger          *logrus.Logger
	successCallback func() // Callback to switch to main layout
	stopRefresh     chan struct{}
	stopOnce        sync.Once
}

// NewVaultProfilesPanel creates a new vault profiles panel
func NewVaultProfilesPanel(config *config.Config, vaultMgr *vault.Manager, app *tview.Application, logger *logrus.Logger) *VaultProfilesPanel {
	return &VaultProfilesPanel{
		config:      config,
		vaultMgr:    vaultMgr,
		app:         app,
		logger:      logger,
		stopRefresh: make(chan struct{}),
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
	if err := vpp.loadProfiles(); err != nil {
		return err
	}

	// Start a background refresher
	go vpp.startRefresher()

	return nil
}

// StopRefresher stops the background refresher
func (vpp *VaultProfilesPanel) StopRefresher() {
	vpp.stopOnce.Do(func() {
		if vpp.stopRefresh != nil {
			close(vpp.stopRefresh)
		}
	})
}

// startRefresher starts a background goroutine to refresh the profiles list
func (vpp *VaultProfilesPanel) startRefresher() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			vpp.app.QueueUpdateDraw(func() {
				// Refresh the UI to show latest statuses
				vpp.Refresh()

				// Stop refreshing if all connections have been resolved
				if !vpp.hasConnectingProfiles() {
					vpp.StopRefresher()
				}
			})
		case <-vpp.stopRefresh:
			return
		}
	}
}

// hasConnectingProfiles checks if there are any profiles that are in the process of connecting
func (vpp *VaultProfilesPanel) hasConnectingProfiles() bool {
	vaults := vpp.vaultMgr.ListVaults()
	for _, vaultName := range vaults {
		status, err := vpp.vaultMgr.GetConnectionStatus(vaultName)
		if err != nil {
			continue
		}
		if status.Connecting {
			return true
		}
	}
	return false
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
				// Manually refresh connections
				vpp.logger.Info("Manual refresh triggered")
				vpp.vaultMgr.GetConnectionManager().RefreshAllConnections()
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
	// Store the current selection
	currentItem := vpp.list.GetCurrentItem()

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

	// Restore selection
	if vpp.list.GetItemCount() > 0 {
		if currentItem >= vpp.list.GetItemCount() {
			vpp.list.SetCurrentItem(vpp.list.GetItemCount() - 1)
		} else {
			vpp.list.SetCurrentItem(currentItem)
		}
	}

	return nil
}

// formatVaultDisplay formats the vault display text with status
func (vpp *VaultProfilesPanel) formatVaultDisplay(name string, status *vault.ConnectionStatus) string {
	var statusIcon string
	var statusText string

	if status.Connecting {
		statusIcon = "⏳"
		statusText = "Connecting"
	} else if status.Connected {
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
	vpp.StopRefresher()

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
	vpp.StopRefresher()

	vpp.logger.Infof("Switched to vault: %s", vaultName)

	// Manually refresh the specific connection
	vpp.vaultMgr.GetConnectionManager().RefreshConnectionStatus(vaultName)

	// Check the connection status
	status, err := vpp.vaultMgr.GetConnectionStatus(vaultName)
	if err != nil {
		vpp.logger.Errorf("Failed to get connection status for '%s': %v", vaultName, err)
		vpp.Refresh()
		return
	}

	// Allow switching if the vault is connected, even if sealed
	// The user might want to unseal it or view its status
	if status.Connected {
		// We have a connection, switch to main layout
		if vpp.successCallback != nil {
			vpp.successCallback()
		}
	} else {
		// Not connected yet, refresh the display
		vpp.logger.Warnf("Vault '%s' is not connected yet (status: %+v)", vaultName, status)
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
		// Split by whitespace to separate icon from vault name
		firstPart := strings.TrimSpace(parts[0])
		fields := strings.Fields(firstPart)

		// The first field is the emoji icon, the second is the vault name
		if len(fields) >= 2 {
			return fields[1]
		}
	}
	return ""
}

// Refresh refreshes the profiles display
func (vpp *VaultProfilesPanel) Refresh() {
	// Reload profiles
	if err := vpp.loadProfiles(); err != nil {
		vpp.logger.Errorf("Failed to refresh profiles: %v", err)
	}
}

// GetPrimitive returns the underlying tview primitive
func (vpp *VaultProfilesPanel) GetPrimitive() tview.Primitive {
	return vpp.list
}
