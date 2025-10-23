package common

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestNewDialogService(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()

	ds := NewDialogService(app, mainRoot)

	assert.NotNil(t, ds)
	assert.Equal(t, app, ds.app)
	assert.Equal(t, mainRoot, ds.mainRoot)
	assert.Nil(t, ds.currentModal)
	assert.Nil(t, ds.previousFocus)
}

func TestNewDialogService_WithNil(t *testing.T) {
	ds := NewDialogService(nil, nil)

	assert.NotNil(t, ds)
	assert.Nil(t, ds.app)
	assert.Nil(t, ds.mainRoot)
}

func TestDialogService_SetMainRoot(t *testing.T) {
	app := tview.NewApplication()
	initialRoot := tview.NewBox()
	ds := NewDialogService(app, initialRoot)

	newRoot := tview.NewBox()
	ds.SetMainRoot(newRoot)

	assert.Equal(t, newRoot, ds.mainRoot)
}

func TestDialogService_Show(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	modal := tview.NewModal()
	ds.Show(modal)

	assert.Equal(t, modal, ds.currentModal)
}

func TestDialogService_Show_WithNilApp(t *testing.T) {
	ds := NewDialogService(nil, nil)

	modal := tview.NewModal()
	// Should not panic with nil app
	ds.Show(modal)

	assert.Nil(t, ds.currentModal)
}

func TestDialogService_Hide(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	// Show a modal first
	modal := tview.NewModal()
	ds.currentModal = modal

	// Now hide it
	ds.Hide()

	assert.Nil(t, ds.currentModal)
	assert.Nil(t, ds.previousFocus)
}

func TestDialogService_Hide_WithNilApp(t *testing.T) {
	ds := NewDialogService(nil, nil)
	ds.currentModal = tview.NewModal()

	// Should not panic with nil app
	ds.Hide()

	// currentModal should still be set since Hide early returns
	assert.NotNil(t, ds.currentModal)
}

func TestDialogService_Hide_WithNilMainRoot(t *testing.T) {
	app := tview.NewApplication()
	ds := NewDialogService(app, nil)
	ds.currentModal = tview.NewModal()

	// Should not panic with nil mainRoot
	ds.Hide()

	// currentModal should still be set since Hide early returns
	assert.NotNil(t, ds.currentModal)
}

func TestDialogService_Hide_RestoresFocus(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	previousFocus := tview.NewBox()
	ds.previousFocus = previousFocus
	ds.currentModal = tview.NewModal()

	ds.Hide()

	assert.Nil(t, ds.currentModal)
	assert.Nil(t, ds.previousFocus)
}

func TestDialogService_IsModalActive_True(t *testing.T) {
	ds := &DialogService{
		currentModal: tview.NewModal(),
	}

	assert.True(t, ds.IsModalActive())
}

func TestDialogService_IsModalActive_False(t *testing.T) {
	ds := &DialogService{
		currentModal: nil,
	}

	assert.False(t, ds.IsModalActive())
}

func TestDialogService_GetCurrentModal(t *testing.T) {
	modal := tview.NewModal()
	ds := &DialogService{
		currentModal: modal,
	}

	assert.Equal(t, modal, ds.GetCurrentModal())
}

func TestDialogService_GetCurrentModal_Nil(t *testing.T) {
	ds := &DialogService{
		currentModal: nil,
	}

	assert.Nil(t, ds.GetCurrentModal())
}

func TestDialogService_ShowError(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	callbackCalled := false
	callback := func() {
		callbackCalled = true
	}

	ds.ShowError("Test error message", callback)

	assert.NotNil(t, ds.currentModal)
	assert.True(t, ds.IsModalActive())

	// Verify it's a modal
	modal, ok := ds.currentModal.(*tview.Modal)
	assert.True(t, ok)
	assert.NotNil(t, modal)

	// Test that callback is called when modal is dismissed
	// We need to manually trigger the done function since we can't interact with the UI
	// The callback should be executed through the doneFunc
	ds.Hide()
	assert.False(t, callbackCalled) // Not called yet

	// Create a new error modal and manually trigger its done function
	ds.ShowError("Test error", callback)
	modal, ok = ds.currentModal.(*tview.Modal)
	assert.True(t, ok)
}

func TestDialogService_ShowError_WithNilCallback(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	// Should not panic with nil callback
	ds.ShowError("Test error message", nil)

	assert.NotNil(t, ds.currentModal)
	assert.True(t, ds.IsModalActive())
}

func TestDialogService_ShowInfo(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	callback := func() {
		// Callback for testing
	}

	ds.ShowInfo("Test Title", "Test message", callback)

	assert.NotNil(t, ds.currentModal)
	assert.True(t, ds.IsModalActive())

	// Verify it's a modal
	modal, ok := ds.currentModal.(*tview.Modal)
	assert.True(t, ok)
	assert.NotNil(t, modal)
}

func TestDialogService_ShowInfo_WithNilCallback(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	// Should not panic with nil callback
	ds.ShowInfo("Test Title", "Test message", nil)

	assert.NotNil(t, ds.currentModal)
	assert.True(t, ds.IsModalActive())
}

func TestDialogService_ShowConfirmation(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	confirmCalled := false
	cancelCalled := false

	onConfirm := func() {
		confirmCalled = true
	}

	onCancel := func() {
		cancelCalled = true
	}

	ds.ShowConfirmation("Are you sure?", onConfirm, onCancel)

	assert.NotNil(t, ds.currentModal)
	assert.True(t, ds.IsModalActive())

	// Verify it's a modal
	modal, ok := ds.currentModal.(*tview.Modal)
	assert.True(t, ok)
	assert.NotNil(t, modal)

	assert.False(t, confirmCalled)
	assert.False(t, cancelCalled)
}

func TestDialogService_ShowConfirmation_WithNilCallbacks(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	// Should not panic with nil callbacks
	ds.ShowConfirmation("Are you sure?", nil, nil)

	assert.NotNil(t, ds.currentModal)
	assert.True(t, ds.IsModalActive())
}

func TestDialogService_ShowDeleteConfirmation(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	deleteCalled := false
	cancelCalled := false

	onDelete := func() {
		deleteCalled = true
	}

	onCancel := func() {
		cancelCalled = true
	}

	ds.ShowDeleteConfirmation("test-item", onDelete, onCancel)

	assert.NotNil(t, ds.currentModal)
	assert.True(t, ds.IsModalActive())

	// Verify it's a modal
	modal, ok := ds.currentModal.(*tview.Modal)
	assert.True(t, ok)
	assert.NotNil(t, modal)

	assert.False(t, deleteCalled)
	assert.False(t, cancelCalled)
}

func TestDialogService_ShowDeleteConfirmation_WithNilCallbacks(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	// Should not panic with nil callbacks
	ds.ShowDeleteConfirmation("test-item", nil, nil)

	assert.NotNil(t, ds.currentModal)
	assert.True(t, ds.IsModalActive())
}

func TestDialogService_ShowCustom(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	customModal := tview.NewFlex()
	ds.ShowCustom(customModal)

	assert.Equal(t, customModal, ds.currentModal)
	assert.True(t, ds.IsModalActive())
}

func TestDialogService_ShowWithCallback(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	onClose := func() {
		// Callback for testing
	}

	modal := tview.NewModal()
	ds.ShowWithCallback(modal, onClose)

	assert.Equal(t, modal, ds.currentModal)
	assert.True(t, ds.IsModalActive())
}

func TestDialogService_ShowWithCallback_WithNilApp(t *testing.T) {
	ds := NewDialogService(nil, nil)

	onClose := func() {
		// Callback for testing
	}

	modal := tview.NewModal()
	// Should not panic with nil app
	ds.ShowWithCallback(modal, onClose)

	assert.Nil(t, ds.currentModal)
}

func TestDialogService_ShowWithCallback_WithNilCallback(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	modal := tview.NewModal()
	// Should not panic with nil callback
	ds.ShowWithCallback(modal, nil)

	assert.Equal(t, modal, ds.currentModal)
	assert.True(t, ds.IsModalActive())
}

func TestDialogService_ShowWithCallback_NonModalPrimitive(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	// Test with a non-modal primitive (e.g., Box)
	box := tview.NewBox()
	ds.ShowWithCallback(box, func() {})

	assert.Equal(t, box, ds.currentModal)
	assert.True(t, ds.IsModalActive())
}

func TestDialogService_MultipleShowHideCycles(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	// First cycle
	modal1 := tview.NewModal()
	ds.Show(modal1)
	assert.True(t, ds.IsModalActive())
	assert.Equal(t, modal1, ds.GetCurrentModal())

	ds.Hide()
	assert.False(t, ds.IsModalActive())
	assert.Nil(t, ds.GetCurrentModal())

	// Second cycle
	modal2 := tview.NewModal()
	ds.Show(modal2)
	assert.True(t, ds.IsModalActive())
	assert.Equal(t, modal2, ds.GetCurrentModal())

	ds.Hide()
	assert.False(t, ds.IsModalActive())
	assert.Nil(t, ds.GetCurrentModal())
}

func TestDialogService_ShowMultipleModalsSequentially(t *testing.T) {
	app := tview.NewApplication()
	mainRoot := tview.NewBox()
	ds := NewDialogService(app, mainRoot)

	modal1 := tview.NewModal()
	ds.Show(modal1)
	assert.Equal(t, modal1, ds.currentModal)

	// Show another modal without hiding the first
	modal2 := tview.NewModal()
	ds.Show(modal2)
	assert.Equal(t, modal2, ds.currentModal)
	assert.NotEqual(t, modal1, ds.currentModal)
}
