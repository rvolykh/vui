package forms

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/ui/common"
	"github.com/sirupsen/logrus"
)

// KeyValueFormConfig defines the template methods for customizing form behavior
type KeyValueFormConfig struct {
	Title          string
	PathLabel      string
	PathEditable   bool
	InitialData    map[string]string
	ShowDeleteKeys bool
	SaveButtonText string
	OnSave         func(path string, data map[string]string)
	OnCancel       func()
}

// KeyValueFormBuilder builds forms for managing key-value pairs using template method pattern
type KeyValueFormBuilder struct {
	app    *tview.Application
	logger *logrus.Logger
	config KeyValueFormConfig

	// Form state
	keyValuePairs map[string]string
	secretPath    string

	// UI components
	container   *tview.Flex
	currentForm *tview.Form
}

// NewKeyValueFormBuilder creates a new form builder
func NewKeyValueFormBuilder(app *tview.Application, logger *logrus.Logger, config KeyValueFormConfig) *KeyValueFormBuilder {
	return &KeyValueFormBuilder{
		app:           app,
		logger:        logger,
		config:        config,
		keyValuePairs: make(map[string]string),
		secretPath:    "",
	}
}

// Build constructs and returns the form primitive
func (fb *KeyValueFormBuilder) Build() tview.Primitive {
	// Initialize state
	if fb.config.InitialData != nil {
		fb.keyValuePairs = make(map[string]string)
		for k, v := range fb.config.InitialData {
			fb.keyValuePairs[k] = v
		}
	}

	// Create container
	fb.container = tview.NewFlex().SetDirection(tview.FlexRow)

	// Build initial form
	fb.rebuildForm()

	return fb.container
}

// SetPath sets the secret path (used for forms where path is set externally)
func (fb *KeyValueFormBuilder) SetPath(path string) {
	fb.secretPath = path
}

// rebuildForm rebuilds the form with current state
func (fb *KeyValueFormBuilder) rebuildForm() {
	form := tview.NewForm()
	fb.currentForm = form

	// Secret path field
	if fb.config.PathLabel != "" {
		pathField := tview.NewInputField().
			SetLabel(fb.config.PathLabel).
			SetText(fb.secretPath).
			SetFieldWidth(50)

		// Make the path field read-only or editable based on config
		if !fb.config.PathEditable || len(fb.keyValuePairs) > 0 {
			pathField.SetDisabled(true)
		} else {
			pathField.SetChangedFunc(func(text string) {
				fb.secretPath = text
			})
		}

		form.AddFormItem(pathField)
	}

	// Show existing key-value pairs
	if len(fb.keyValuePairs) > 0 {
		pairText := fmt.Sprintf("Key-Value Pairs (%d):", len(fb.keyValuePairs))
		form.AddTextView(pairText, "", 0, 1, false, false)

		// Sort keys for consistent display
		keys := make([]string, 0, len(fb.keyValuePairs))
		for k := range fb.keyValuePairs {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := fb.keyValuePairs[k]
			displayValue := v
			if len(displayValue) > 40 {
				displayValue = displayValue[:37] + "..."
			}

			if fb.config.ShowDeleteKeys {
				// For edit mode: show as input field with delete button
				form.AddInputField(k, v, 50, nil, func(text string) {
					fb.keyValuePairs[k] = text
				})
				form.AddButton("Delete "+k, fb.createDeleteKeyHandler(k))
			} else {
				// For create mode: show as read-only text
				form.AddTextView("  • "+k, displayValue, 0, 1, false, false)
			}
		}
	}

	// New key-value pair section
	form.AddTextView("New Pair:", "", 0, 1, false, false)
	form.AddInputField("Key", "", 30, nil, nil).
		AddInputField("Value", "", 50, nil, nil)

	// Buttons
	form.AddButton("Add Key-Value", fb.createAddKeyValueHandler()).
		AddButton(fb.config.SaveButtonText, fb.createSaveHandler()).
		AddButton("Cancel", func() {
			if fb.config.OnCancel != nil {
				fb.config.OnCancel()
			}
		})

	// Set title
	title := fb.config.Title
	if len(fb.keyValuePairs) > 0 {
		title = fmt.Sprintf("%s [%d pairs]", title, len(fb.keyValuePairs))
	}
	form.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft)

	// Apply styling
	common.ApplyFormStyle(form)
	common.ApplyButtonStyle(form)

	// Replace the form in the container
	fb.container.Clear()
	fb.container.AddItem(form, 0, 1, true)
}

// createAddKeyValueHandler creates the handler for adding key-value pairs
func (fb *KeyValueFormBuilder) createAddKeyValueHandler() func() {
	return func() {
		keyField, valueField := fb.findKeyValueFields()
		if keyField == nil || valueField == nil {
			fb.logger.Error("Could not find Key and Value fields")
			return
		}

		key := strings.TrimSpace(keyField.GetText())
		value := strings.TrimSpace(valueField.GetText())

		if key == "" {
			fb.logger.Warn("Key cannot be empty")
			return
		}

		if value == "" {
			fb.logger.Warn("Value cannot be empty")
			return
		}

		// Add to the pairs map
		fb.keyValuePairs[key] = value
		fb.logger.Infof("Added key-value pair: %s", key)

		// Rebuild the form
		fb.app.QueueUpdateDraw(func() {
			fb.rebuildForm()
			if fb.currentForm != nil {
				fb.app.SetFocus(fb.currentForm)
			}
		})
	}
}

// createSaveHandler creates the handler for saving the form
func (fb *KeyValueFormBuilder) createSaveHandler() func() {
	return func() {
		// Auto-add unsaved key-value fields before saving
		keyField, valueField := fb.findKeyValueFields()
		if keyField != nil && valueField != nil {
			key := strings.TrimSpace(keyField.GetText())
			value := strings.TrimSpace(valueField.GetText())

			if key != "" && value != "" {
				fb.keyValuePairs[key] = value
				fb.logger.Infof("Auto-added key-value pair before save: %s", key)
			}
		}

		// Call the save callback
		if fb.config.OnSave != nil {
			fb.config.OnSave(fb.secretPath, fb.keyValuePairs)
		}
	}
}

// createDeleteKeyHandler creates a handler for deleting a specific key
func (fb *KeyValueFormBuilder) createDeleteKeyHandler(key string) func() {
	return func() {
		delete(fb.keyValuePairs, key)
		fb.logger.Infof("Deleted key-value pair: %s", key)

		// Rebuild the form
		fb.app.QueueUpdateDraw(func() {
			fb.rebuildForm()
			if fb.currentForm != nil {
				fb.app.SetFocus(fb.currentForm)
			}
		})
	}
}

// findKeyValueFields finds the Key and Value input fields in the form
func (fb *KeyValueFormBuilder) findKeyValueFields() (*tview.InputField, *tview.InputField) {
	var keyField, valueField *tview.InputField

	if fb.currentForm == nil {
		return nil, nil
	}

	// Scan backward from buttons to find the new Key and Value fields
	for i := fb.currentForm.GetFormItemCount() - 1; i >= 0; i-- {
		item := fb.currentForm.GetFormItem(i)
		if inputField, ok := item.(*tview.InputField); ok {
			label := inputField.GetLabel()
			if label == "Value" && valueField == nil {
				valueField = inputField
			} else if label == "Key" && keyField == nil {
				keyField = inputField
			}
		}
	}

	return keyField, valueField
}
