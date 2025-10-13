package common

import (
	"fmt"

	"github.com/rivo/tview"
)

// DialogService manages modal dialogs, combining creation and lifecycle management
type DialogService struct {
	app           *tview.Application
	mainRoot      tview.Primitive
	currentModal  tview.Primitive
	previousFocus tview.Primitive
}

// NewDialogService creates a new dialog service
func NewDialogService(app *tview.Application, mainRoot tview.Primitive) *DialogService {
	return &DialogService{
		app:      app,
		mainRoot: mainRoot,
	}
}

// SetMainRoot updates the main root primitive
func (ds *DialogService) SetMainRoot(mainRoot tview.Primitive) {
	ds.mainRoot = mainRoot
}

// Show displays a modal
func (ds *DialogService) Show(modal tview.Primitive) {
	if ds.app == nil {
		return
	}

	ds.currentModal = modal
	ds.app.SetRoot(modal, true)
}

// Hide hides the current modal and returns to the main view
func (ds *DialogService) Hide() {
	if ds.app == nil || ds.mainRoot == nil {
		return
	}

	ds.currentModal = nil
	ds.app.SetRoot(ds.mainRoot, true)

	// Restore focus if needed
	if ds.previousFocus != nil {
		ds.app.SetFocus(ds.previousFocus)
		ds.previousFocus = nil
	}
}

// IsModalActive returns true if a modal is currently displayed
func (ds *DialogService) IsModalActive() bool {
	return ds.currentModal != nil
}

// GetCurrentModal returns the current modal (or nil if none)
func (ds *DialogService) GetCurrentModal() tview.Primitive {
	return ds.currentModal
}

// ShowError displays an error dialog
func (ds *DialogService) ShowError(message string, callback func()) {
	modal := NewModalBuilder().
		SetText(fmt.Sprintf("[red]Error[white]\n\n%s", message)).
		AddButton("OK").
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ds.Hide()
			if callback != nil {
				callback()
			}
		}).
		Build()

	ds.Show(modal)
}

// ShowInfo displays an info dialog
func (ds *DialogService) ShowInfo(title, message string, callback func()) {
	modal := NewModalBuilder().
		SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s", title, message)).
		AddButton("OK").
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ds.Hide()
			if callback != nil {
				callback()
			}
		}).
		Build()

	ds.Show(modal)
}

// ShowConfirmation displays a confirmation dialog
func (ds *DialogService) ShowConfirmation(message string, onConfirm func(), onCancel func()) {
	modal := NewModalBuilder().
		SetText(message).
		AddButtons("Cancel", "Confirm").
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ds.Hide()
			if buttonLabel == "Confirm" && onConfirm != nil {
				onConfirm()
			} else if onCancel != nil {
				onCancel()
			}
		}).
		Build()

	ds.Show(modal)
}

// ShowDeleteConfirmation displays a delete confirmation dialog with warning styling
func (ds *DialogService) ShowDeleteConfirmation(itemName string, onDelete func(), onCancel func()) {
	modal := NewModalBuilder().
		SetText(fmt.Sprintf("Are you sure you want to delete:\n\n[::b]%s[::-]\n\nThis action cannot be undone.", itemName)).
		AddButtons("Cancel", "Delete").
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ds.Hide()
			if buttonLabel == "Delete" && onDelete != nil {
				onDelete()
			} else if onCancel != nil {
				onCancel()
			}
		}).
		UseDeleteStyle().
		Build()

	ds.Show(modal)
}

// ShowCustom displays a custom modal (for forms or complex dialogs)
func (ds *DialogService) ShowCustom(modal tview.Primitive) {
	ds.Show(modal)
}

// ShowWithCallback displays a modal and calls the callback when it's closed
func (ds *DialogService) ShowWithCallback(modal tview.Primitive, onClose func()) {
	if ds.app == nil {
		return
	}

	ds.currentModal = modal

	// For modals, wrap the done function to include our callback
	if m, ok := modal.(*tview.Modal); ok {
		m.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			ds.Hide()
			if onClose != nil {
				onClose()
			}
		})
	}

	ds.app.SetRoot(modal, true)
}
