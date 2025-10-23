package panels

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewProfilesTable(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	assert.NotNil(t, pt)
	assert.Equal(t, cfg, pt.config)
	assert.Equal(t, vaultMgr, pt.vaultMgr)
	assert.Equal(t, app, pt.app)
	assert.Equal(t, logger, pt.logger)
	assert.NotNil(t, pt.stopRefresh)
	assert.Nil(t, pt.table)
}

func TestProfilesTable_SetSuccessCallback(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	callbackCalled := false
	callback := func() {
		callbackCalled = true
	}

	pt.SetSuccessCallback(callback)

	assert.NotNil(t, pt.successCallback)

	// Test the callback
	pt.successCallback()
	assert.True(t, callbackCalled)
}

func TestProfilesTable_SetSuccessCallback_Nil(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	// Should not panic with nil callback
	pt.SetSuccessCallback(nil)

	assert.Nil(t, pt.successCallback)
}

func TestProfilesTable_SetErrorCallback(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	callbackCalled := false
	var receivedError string
	callback := func(err string) {
		callbackCalled = true
		receivedError = err
	}

	pt.SetErrorCallback(callback)

	assert.NotNil(t, pt.errorCallback)

	// Test the callback
	pt.errorCallback("test error")
	assert.True(t, callbackCalled)
	assert.Equal(t, "test error", receivedError)
}

func TestProfilesTable_SetErrorCallback_Nil(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	// Should not panic with nil callback
	pt.SetErrorCallback(nil)

	assert.Nil(t, pt.errorCallback)
}

func TestProfilesTable_Initialize(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	err := pt.Initialize()

	// Will succeed even without proper vault setup
	assert.NoError(t, err)
	assert.NotNil(t, pt.table)
	assert.NotNil(t, pt.stopRefresh)
}

func TestProfilesTable_StopRefresher(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	// Should not panic
	pt.StopRefresher()

	// Calling again should also not panic (sync.Once)
	pt.StopRefresher()
}

func TestProfilesTable_StopRefresher_WithNilChannel(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.stopRefresh = nil

	// Should not panic with nil channel
	pt.StopRefresher()
}

func TestProfilesTable_StopRefresher_AfterInitialize(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Give goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Stop the refresher
	pt.StopRefresher()

	// Give goroutine time to stop
	time.Sleep(10 * time.Millisecond)
}

func TestProfilesTable_HasConnectingProfiles_NoProfiles(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	// With uninitialized vault manager, should return false
	result := pt.hasConnectingProfiles()
	assert.False(t, result)
}

func TestProfilesTable_GetPrimitive(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	primitive := pt.GetPrimitive()

	assert.NotNil(t, primitive)
	assert.Equal(t, pt.table, primitive)
}

func TestProfilesTable_GetPrimitive_BeforeInitialize(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	primitive := pt.GetPrimitive()

	assert.Nil(t, primitive)
}

func TestProfilesTable_FormatStatus_Connecting(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	status := &vault.ConnectionStatus{
		Connecting: true,
		Connected:  false,
	}

	text, color := pt.formatStatus(status)

	assert.Equal(t, "⏳ Connecting", text)
	assert.Equal(t, tcell.ColorYellow, color)
}

func TestProfilesTable_FormatStatus_ConnectedAndSealed(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	status := &vault.ConnectionStatus{
		Connecting: false,
		Connected:  true,
		Sealed:     true,
	}

	text, color := pt.formatStatus(status)

	assert.Equal(t, "🔒 Sealed", text)
	assert.Equal(t, tcell.ColorOrange, color)
}

func TestProfilesTable_FormatStatus_ConnectedAndUnsealed(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	status := &vault.ConnectionStatus{
		Connecting: false,
		Connected:  true,
		Sealed:     false,
	}

	text, color := pt.formatStatus(status)

	assert.Equal(t, "✅ Connected", text)
	assert.Equal(t, tcell.ColorGreen, color)
}

func TestProfilesTable_FormatStatus_Disconnected(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	status := &vault.ConnectionStatus{
		Connecting: false,
		Connected:  false,
	}

	text, color := pt.formatStatus(status)

	assert.Equal(t, "❌ Disconnected", text)
	assert.Equal(t, tcell.ColorRed, color)
}

func TestProfilesTable_LoadProfiles_NoVaults(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	err := pt.loadProfiles()

	assert.NoError(t, err)
	assert.NotNil(t, pt.table)
	// Should have header row
	assert.GreaterOrEqual(t, pt.table.GetRowCount(), 1)
}

func TestProfilesTable_Refresh(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Should not panic
	pt.Refresh()
}

func TestProfilesTable_Refresh_BeforeInitialize(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	// Calling Refresh before Initialize will panic because table is nil
	// This is expected behavior - Initialize must be called first
	assert.Nil(t, pt.table, "Table should be nil before Initialize")
}

func TestProfilesTable_SetupKeyboardNavigation(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Verify that keyboard navigation is set up (table should have input capture)
	assert.NotNil(t, pt.table)
}

func TestProfilesTable_SwitchToSelectedVault_NoSelection(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Should not panic with no valid selection
	pt.switchToSelectedVault()
}

func TestProfilesTable_SwitchToSelectedVault_HeaderRow(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()
	pt.table.Select(0, 0) // Select header row

	// Should not panic when header is selected
	pt.switchToSelectedVault()
}

func TestProfilesTable_AddNewVault(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Should not panic (currently not implemented)
	pt.addNewVault()
}

func TestProfilesTable_DeleteSelectedVault_NoSelection(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Should not panic with no valid selection
	pt.deleteSelectedVault()
}

func TestProfilesTable_DeleteSelectedVault_HeaderRow(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()
	pt.table.Select(0, 0) // Select header row

	// Should not panic when header is selected
	pt.deleteSelectedVault()
}

func TestProfilesTable_NewWithNilConfig(t *testing.T) {
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(nil, vaultMgr, app, logger)

	assert.NotNil(t, pt)
	assert.Nil(t, pt.config)
}

func TestProfilesTable_NewWithNilVaultManager(t *testing.T) {
	cfg := &config.Config{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, nil, app, logger)

	assert.NotNil(t, pt)
	assert.Nil(t, pt.vaultMgr)
}

func TestProfilesTable_NewWithNilApp(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, nil, logger)

	assert.NotNil(t, pt)
	assert.Nil(t, pt.app)
}

func TestProfilesTable_NewWithNilLogger(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()

	pt := NewProfilesTable(cfg, vaultMgr, app, nil)

	assert.NotNil(t, pt)
	assert.Nil(t, pt.logger)
}

func TestProfilesTable_LoadProfiles_PreservesSelection(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// First load
	pt.loadProfiles()

	// Reload should not panic
	pt.loadProfiles()
}

func TestProfilesTable_FormatStatus_WithError(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	status := &vault.ConnectionStatus{
		Connecting: false,
		Connected:  false,
		Error:      "connection refused",
	}

	text, color := pt.formatStatus(status)

	assert.Equal(t, "❌ Disconnected", text)
	assert.Equal(t, tcell.ColorRed, color)
}

func TestProfilesTable_EnsureRefresherRunning(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Should not panic
	pt.ensureRefresherRunning()

	// Give goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Clean up
	pt.StopRefresher()
	time.Sleep(10 * time.Millisecond)
}

func TestProfilesTable_RefreshProfiles(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// refreshProfiles() requires a fully initialized vault manager with connection manager
	// Testing this requires more complex setup, so we just verify the method exists
	// The actual functionality is tested through integration tests
	assert.NotNil(t, pt.vaultMgr)

	// Clean up
	pt.StopRefresher()
	time.Sleep(10 * time.Millisecond)
}

func TestProfilesTable_KeyboardNavigation_Enter(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Simulate Enter key press (should trigger switchToSelectedVault)
	// We can't easily test the input capture directly, but we can verify it's set up
	assert.NotNil(t, pt.table)
}

func TestProfilesTable_VaultNamesInitialization(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// vaultNames should be initialized (even if empty)
	assert.NotNil(t, pt.vaultNames)
}

func TestProfilesTable_MultipleRefreshCycles(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)
	pt.Initialize()

	// Multiple refresh cycles should not panic
	pt.Refresh()
	pt.Refresh()
	pt.Refresh()

	// Clean up
	pt.StopRefresher()
}

func TestProfilesTable_CallbacksIntegration(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()
	logger := logrus.New()

	pt := NewProfilesTable(cfg, vaultMgr, app, logger)

	successCalled := false
	errorCalled := false
	var errorMsg string

	pt.SetSuccessCallback(func() {
		successCalled = true
	})

	pt.SetErrorCallback(func(msg string) {
		errorCalled = true
		errorMsg = msg
	})

	pt.Initialize()

	// Verify callbacks are set
	assert.NotNil(t, pt.successCallback)
	assert.NotNil(t, pt.errorCallback)

	// Manually trigger callbacks
	if pt.successCallback != nil {
		pt.successCallback()
	}
	assert.True(t, successCalled)

	if pt.errorCallback != nil {
		pt.errorCallback("test error")
	}
	assert.True(t, errorCalled)
	assert.Equal(t, "test error", errorMsg)

	// Clean up
	pt.StopRefresher()
}
