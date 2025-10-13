package common

import (
	"fmt"

	"github.com/rivo/tview"
)

// ModalBuilder provides a fluent API for creating modals
type ModalBuilder struct {
	text        string
	buttons     []string
	doneFunc    func(buttonIndex int, buttonLabel string)
	deleteStyle bool
}

// NewModalBuilder creates a new modal builder
func NewModalBuilder() *ModalBuilder {
	return &ModalBuilder{
		buttons: []string{},
	}
}

// SetText sets the modal text
func (mb *ModalBuilder) SetText(text string) *ModalBuilder {
	mb.text = text
	return mb
}

// AddButton adds a button to the modal
func (mb *ModalBuilder) AddButton(label string) *ModalBuilder {
	mb.buttons = append(mb.buttons, label)
	return mb
}

// AddButtons adds multiple buttons to the modal
func (mb *ModalBuilder) AddButtons(labels ...string) *ModalBuilder {
	mb.buttons = append(mb.buttons, labels...)
	return mb
}

// SetDoneFunc sets the done callback
func (mb *ModalBuilder) SetDoneFunc(fn func(buttonIndex int, buttonLabel string)) *ModalBuilder {
	mb.doneFunc = fn
	return mb
}

// UseDeleteStyle applies destructive styling (for delete confirmations)
func (mb *ModalBuilder) UseDeleteStyle() *ModalBuilder {
	mb.deleteStyle = true
	return mb
}

// Build constructs and returns the modal
func (mb *ModalBuilder) Build() *tview.Modal {
	modal := tview.NewModal().
		SetText(mb.text).
		AddButtons(mb.buttons)

	if mb.doneFunc != nil {
		modal.SetDoneFunc(mb.doneFunc)
	}

	// Apply styling
	if mb.deleteStyle {
		style := GetDeleteButtonStyle()
		modal.SetButtonBackgroundColor(style.Background).
			SetButtonTextColor(style.Text).
			SetButtonActivatedStyle(style.ActivatedStyle)
	} else {
		ApplyModalStyle(modal)
	}

	return modal
}

// ErrorModal creates a standard error modal
func ErrorModal(message string, callback func()) *tview.Modal {
	return NewModalBuilder().
		SetText(fmt.Sprintf("[red]Error[white]\n\n%s", message)).
		AddButton("OK").
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if callback != nil {
				callback()
			}
		}).
		Build()
}

// ConfirmationModal creates a standard confirmation modal
func ConfirmationModal(message string, onConfirm func(), onCancel func()) *tview.Modal {
	return NewModalBuilder().
		SetText(message).
		AddButtons("Cancel", "Confirm").
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Confirm" && onConfirm != nil {
				onConfirm()
			} else if onCancel != nil {
				onCancel()
			}
		}).
		Build()
}

// DeleteConfirmationModal creates a delete confirmation modal with warning styling
func DeleteConfirmationModal(itemName string, onDelete func(), onCancel func()) *tview.Modal {
	return NewModalBuilder().
		SetText(fmt.Sprintf("Are you sure you want to delete:\n\n[::b]%s[::-]\n\nThis action cannot be undone.", itemName)).
		AddButtons("Cancel", "Delete").
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" && onDelete != nil {
				onDelete()
			} else if onCancel != nil {
				onCancel()
			}
		}).
		UseDeleteStyle().
		Build()
}

// InfoModal creates a standard info modal
func InfoModal(title, message string, callback func()) *tview.Modal {
	return NewModalBuilder().
		SetText(fmt.Sprintf("[yellow]%s[white]\n\n%s", title, message)).
		AddButton("OK").
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if callback != nil {
				callback()
			}
		}).
		Build()
}
