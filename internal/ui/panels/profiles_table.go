package panels

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/backend"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
)

// ProfilesTable displays vault profiles with connection status
type ProfilesTable struct {
	config          *config.Config
	interactor      backend.Interactor
	table           *tview.Table
	app             *tview.Application
	logger          *logrus.Logger
	successCallback func()       // Callback to switch to main layout
	errorCallback   func(string) // Callback to show error message
	stopRefresh     chan struct{}
	stopOnce        sync.Once
	vaultNames      []string // Sorted list of vault names for selection tracking
}

// NewProfilesTable creates a new vault profiles panel
func NewProfilesTable(config *config.Config, interactor backend.Interactor, app *tview.Application, logger *logrus.Logger) *ProfilesTable {
	return &ProfilesTable{
		config:      config,
		interactor:  interactor,
		app:         app,
		logger:      logger,
		stopRefresh: make(chan struct{}),
	}
}

// SetSuccessCallback sets the callback to be called when a vault is successfully connected
func (vpp *ProfilesTable) SetSuccessCallback(callback func()) {
	vpp.successCallback = callback
}

// SetErrorCallback sets the callback to be called when an error occurs
func (vpp *ProfilesTable) SetErrorCallback(callback func(string)) {
	vpp.errorCallback = callback
}

// Initialize initializes the vault profiles panel
func (vpp *ProfilesTable) Initialize() error {
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
func (vpp *ProfilesTable) StopRefresher() {
	vpp.stopOnce.Do(func() {
		if vpp.stopRefresh != nil {
			close(vpp.stopRefresh)
		}
	})
}

// startRefresher starts a background goroutine to refresh the profiles list
func (vpp *ProfilesTable) startRefresher() {
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
func (vpp *ProfilesTable) hasConnectingProfiles() bool {
	profiles := vpp.interactor.Profiles().ListConnections()
	for _, name := range profiles {
		status, err := vpp.interactor.Profiles().GetConnectionStatus(name)
		if err != nil {
			continue
		}
		if status.Status == models.StatusConnecting {
			return true
		}
	}
	return false
}

// setupKeyboardNavigation sets up keyboard navigation
func (vpp *ProfilesTable) setupKeyboardNavigation() {
	vpp.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			// Switch to selected vault
			vpp.switchToSelectedVault()
			return nil
		case tcell.KeyF5:
			// Refresh connections (F5)
			vpp.logger.Info("Manual refresh triggered (F5)")
			vpp.refreshProfiles()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'r':
				// Manually refresh connections
				vpp.logger.Info("Manual refresh triggered")
				vpp.refreshProfiles()
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
func (vpp *ProfilesTable) loadProfiles() error {
	// Store the current selection (row number, accounting for header)
	currentRow, _ := vpp.table.GetSelection()

	// Clear existing items
	vpp.table.Clear()

	// Get available profiles
	profiles := vpp.interactor.Profiles().ListConnections()

	// Sort profiles by name (ascending)
	sort.Strings(profiles)
	vpp.vaultNames = profiles

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

	if len(profiles) == 0 {
		// Show a message when no profiles are configured
		cell := tview.NewTableCell("No profiles configured").
			SetAlign(tview.AlignCenter).
			SetExpansion(1)
		vpp.table.SetCell(1, 0, cell)
		return nil
	}

	// Add each vault profile as a row
	row := 1
	for _, name := range profiles {
		// Get connection status
		status, err := vpp.interactor.Profiles().GetConnectionStatus(name)
		if err != nil {
			vpp.logger.Warnf("Failed to get status for profile '%s': %v", name, err)
			continue
		}

		// Column 0: Name
		nameCell := tview.NewTableCell(name).
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
func (vpp *ProfilesTable) formatStatus(status *models.ConnectionStatus) (string, tcell.Color) {
	if status.Status == models.StatusConnecting {
		return "⏳ Connecting", tcell.ColorYellow
	} else if status.Status == models.StatusConnected {
		return "✅ Connected", tcell.ColorGreen
	} else if status.Status == models.StatusSealed {
		return "🔒 Sealed", tcell.ColorOrange
	} else {
		return "❌ Disconnected", tcell.ColorRed
	}
}

// switchToSelectedVault switches to the currently selected vault
func (vpp *ProfilesTable) switchToSelectedVault() {
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
func (vpp *ProfilesTable) switchToVault(name string) {
	if err := vpp.interactor.Profiles().SwitchProfile(name); err != nil {
		vpp.logger.Errorf("Failed to switch to profile '%s': %v", name, err)
		// Show error dialog if callback is set
		if vpp.errorCallback != nil {
			vpp.errorCallback(fmt.Sprintf("Failed to connect to profile '%s':\n\n%v", name, err))
		}
		vpp.Refresh()
		return
	}
	vpp.StopRefresher()

	vpp.logger.Infof("Switched to profile: %s", name)

	// Manually refresh the specific connection
	vpp.interactor.Profiles().RefreshConnection(name)

	// Check the connection status
	status, err := vpp.interactor.Profiles().GetConnectionStatus(name)
	if err != nil {
		vpp.logger.Errorf("Failed to get connection status for '%s': %v", name, err)
		if vpp.errorCallback != nil {
			vpp.errorCallback(fmt.Sprintf("Failed to get connection status for '%s':\n\n%v", name, err))
		}
		vpp.Refresh()
		return
	}

	// Allow switching if the vault is connected, even if sealed
	// The user might want to unseal it or view its status
	if status.Status == models.StatusConnected {
		// We have a connection, switch to main layout
		if vpp.successCallback != nil {
			vpp.successCallback()
		}
	} else {
		// Not connected yet, show error
		vpp.logger.Warnf("Profile '%s' is not connected yet (status: %+v)", name, status)
		if vpp.errorCallback != nil {
			errorMsg := fmt.Sprintf("Cannot connect to profile '%s'", name)
			if status.Error != "" {
				errorMsg = fmt.Sprintf("Cannot connect to profile '%s':\n\n%s", name, status.Error)
			}
			vpp.errorCallback(errorMsg)
		}
		vpp.Refresh()
	}
}

// addNewVault adds a new vault profile
func (vpp *ProfilesTable) addNewVault() {
	// This would show a form to add a new vault
	// For now, just log the action
	vpp.logger.Info("Add new vault (not implemented yet)")
}

// deleteSelectedVault deletes the selected vault profile
func (vpp *ProfilesTable) deleteSelectedVault() {
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
func (vpp *ProfilesTable) Refresh() {
	// Reload profiles
	if err := vpp.loadProfiles(); err != nil {
		vpp.logger.Errorf("Failed to refresh profiles: %v", err)
	}
}

// refreshProfiles reloads the configuration and refreshes all connections
func (vpp *ProfilesTable) refreshProfiles() {
	vpp.logger.Info("Reloading configuration and refreshing connections...")

	// Run the reload in a goroutine to avoid blocking the UI thread
	go func() {
		// Set all connections to "Connecting" state immediately
		vpp.interactor.Profiles().ResetConnections()

		// Refresh the display to show "Connecting" state
		vpp.app.QueueUpdateDraw(func() {
			vpp.Refresh()
		})

		// Reload configuration from disk (this will test connections asynchronously)
		if err := vpp.interactor.Profiles().ReloadConfiguration(); err != nil {
			vpp.logger.Errorf("Failed to reload configuration: %v", err)
			vpp.app.QueueUpdateDraw(func() {
				if vpp.errorCallback != nil {
					vpp.errorCallback(fmt.Sprintf("Failed to reload configuration:\n\n%v", err))
				}
			})
			return
		}

		// Ensure the background refresher is running to update UI as connections complete
		// We need to restart it since we now have connecting profiles
		vpp.ensureRefresherRunning()

		vpp.logger.Info("Configuration reload initiated, testing connections...")
	}()
}

// ensureRefresherRunning ensures the background refresher is running
func (vpp *ProfilesTable) ensureRefresherRunning() {
	// Try to stop existing refresher
	vpp.stopOnce.Do(func() {
		if vpp.stopRefresh != nil {
			close(vpp.stopRefresh)
		}
	})

	// Create new refresher
	vpp.stopRefresh = make(chan struct{})
	vpp.stopOnce = sync.Once{}
	go vpp.startRefresher()

	vpp.logger.Debug("Background refresher started")
}

// GetPrimitive returns the underlying tview primitive
func (vpp *ProfilesTable) GetPrimitive() tview.Primitive {
	return vpp.table
}
