package ui

import (
	"sort"
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
	table           *tview.Table
	app             *tview.Application
	logger          *logrus.Logger
	successCallback func() // Callback to switch to main layout
	stopRefresh     chan struct{}
	stopOnce        sync.Once
	vaultNames      []string // Sorted list of vault names for selection tracking
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
	vpp.table = tview.NewTable()

	// Set up the table appearance
	vpp.table.SetBorder(true).
		SetTitle("Vault Profiles").
		SetTitleAlign(tview.AlignLeft)
	vpp.table.SetSelectable(true, false).
		SetFixed(1, 0) // Fix the header row

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
	vpp.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
	// Store the current selection (row number, accounting for header)
	currentRow, _ := vpp.table.GetSelection()

	// Clear existing items
	vpp.table.Clear()

	// Get available vaults
	vaults := vpp.vaultMgr.ListVaults()

	// Sort vaults by name (ascending)
	sort.Strings(vaults)
	vpp.vaultNames = vaults

	// Create header row
	headers := []string{"Name", "Address", "Status", "NOTE"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold)
		vpp.table.SetCell(0, col, cell)
	}

	if len(vaults) == 0 {
		// Show a message when no vaults are configured
		cell := tview.NewTableCell("No vault profiles configured").
			SetAlign(tview.AlignCenter).
			SetExpansion(1)
		vpp.table.SetCell(1, 0, cell)
		return nil
	}

	// Add each vault profile as a row
	row := 1
	for _, vaultName := range vaults {
		// Get connection status
		status, err := vpp.vaultMgr.GetConnectionStatus(vaultName)
		if err != nil {
			vpp.logger.Warnf("Failed to get status for vault '%s': %v", vaultName, err)
			continue
		}

		// Column 0: Name
		nameCell := tview.NewTableCell(vaultName).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft)
		vpp.table.SetCell(row, 0, nameCell)

		// Column 1: Address
		addressCell := tview.NewTableCell(status.Address).
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignLeft)
		vpp.table.SetCell(row, 1, addressCell)

		// Column 2: Status
		statusText, statusColor := vpp.formatStatus(status)
		statusCell := tview.NewTableCell(statusText).
			SetTextColor(statusColor).
			SetAlign(tview.AlignCenter)
		vpp.table.SetCell(row, 2, statusCell)

		// Column 3: NOTE (error message)
		noteText := ""
		if status.Error != "" {
			noteText = status.Error
		}
		noteCell := tview.NewTableCell(noteText).
			SetTextColor(tcell.ColorRed).
			SetAlign(tview.AlignLeft).
			SetMaxWidth(0). // Allow wrapping
			SetExpansion(1) // Allow expansion for long text
		vpp.table.SetCell(row, 3, noteCell)

		row++
	}

	// Restore selection
	rowCount := vpp.table.GetRowCount()
	if rowCount > 1 {
		if currentRow >= rowCount {
			vpp.table.Select(rowCount-1, 0)
		} else if currentRow < 1 {
			vpp.table.Select(1, 0) // Skip header
		} else {
			vpp.table.Select(currentRow, 0)
		}
	}

	return nil
}

// formatStatus formats the status text and color
func (vpp *VaultProfilesPanel) formatStatus(status *vault.ConnectionStatus) (string, tcell.Color) {
	if status.Connecting {
		return "⏳ Connecting", tcell.ColorYellow
	} else if status.Connected {
		if status.Sealed {
			return "🔒 Sealed", tcell.ColorOrange
		} else {
			return "✅ Connected", tcell.ColorGreen
		}
	} else {
		return "❌ Disconnected", tcell.ColorRed
	}
}

// switchToSelectedVault switches to the currently selected vault
func (vpp *VaultProfilesPanel) switchToSelectedVault() {
	row, _ := vpp.table.GetSelection()
	// Row 0 is header, data starts at row 1
	if row < 1 || row > len(vpp.vaultNames) {
		return
	}
	vpp.StopRefresher()

	// Get the vault name from the sorted list (row-1 because row 0 is header)
	vaultName := vpp.vaultNames[row-1]
	vpp.switchToVault(vaultName)
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
	row, _ := vpp.table.GetSelection()
	// Row 0 is header, data starts at row 1
	if row < 1 || row > len(vpp.vaultNames) {
		return
	}

	// Get the vault name from the sorted list (row-1 because row 0 is header)
	vaultName := vpp.vaultNames[row-1]
	vpp.logger.Infof("Delete vault: %s (not implemented yet)", vaultName)
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
	return vpp.table
}
