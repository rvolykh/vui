package forms

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/ui/common"
	"github.com/rvolykh/vui/internal/utils"
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
		maps.Copy(fb.keyValuePairs, fb.config.InitialData)
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
	fb.logger.Info("rebuildForm: Starting")
	form := tview.NewForm()
	fb.currentForm = form
	fb.logger.Info("rebuildForm: Created new form")

	// Secret path field
	if fb.config.PathLabel != "" {
		fb.logger.Infof("rebuildForm: Adding path field, label='%s'", fb.config.PathLabel)
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
		fb.logger.Info("rebuildForm: Path field added")
	}

	// Show existing key-value pairs
	if len(fb.keyValuePairs) > 0 {
		fb.logger.Infof("rebuildForm: Processing %d existing key-value pairs", len(fb.keyValuePairs))
		pairText := fmt.Sprintf("Key-Value Pairs (%d):", len(fb.keyValuePairs))
		form.AddTextView(pairText, "", 0, 1, false, false)

		// Sort keys for consistent display
		keys := make([]string, 0, len(fb.keyValuePairs))
		for k := range fb.keyValuePairs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fb.logger.Infof("rebuildForm: Sorted keys: %v", keys)

		for idx, k := range keys {
			fb.logger.Infof("rebuildForm: Processing pair %d/%d: key='%s'", idx+1, len(keys), k)
			v := fb.keyValuePairs[k]

			// Display key label - show "(no key)" for empty keys
			displayKey := utils.Coalesce(k, "(no key)")

			if fb.config.ShowDeleteKeys {
				fb.logger.Infof("rebuildForm: Adding delete button for key='%s'", k)
				// For edit mode: show as input field with delete button
				// Create a copy of k for the closure
				key := k

				valueField := tview.NewTextArea().SetLabel(displayKey).SetText(v, true)
				form.AddFormItem(valueField)
				form.AddButton("Delete "+displayKey, fb.createDeleteKeyHandler(key))
			} else {
				displayValue := strings.ReplaceAll(v, "\n", "\\n")
				if len(displayValue) > 40 {
					displayValue = displayValue[:37] + "..."
				}

				fb.logger.Infof("rebuildForm: Adding read-only display for key='%s'", k)
				// For create mode: show as read-only text
				form.AddTextView("  • "+displayKey, displayValue, 0, 1, false, false)
			}
		}
		fb.logger.Info("rebuildForm: Finished processing existing pairs")
	}

	// New key-value pair section
	fb.logger.Info("rebuildForm: Adding new pair section")
	form.AddTextView("New Pair:", "", 0, 1, false, false)

	// Add Key field
	fb.logger.Info("rebuildForm: Adding Key field")
	form.AddInputField("Key", "", 30, nil, nil)

	// Add Value field with multiline support
	fb.logger.Info("rebuildForm: Adding Value field (TextArea)")
	valueField := tview.NewTextArea().SetLabel("Value").SetText("", true)
	form.AddFormItem(valueField)

	// Buttons
	fb.logger.Info("rebuildForm: Adding buttons")
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
	fb.logger.Infof("rebuildForm: Setting title='%s'", title)
	form.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft)

	// Apply styling
	fb.logger.Info("rebuildForm: Applying styles")
	common.ApplyFormStyle(form)
	common.ApplyButtonStyle(form)

	// Replace the form in the container
	fb.logger.Info("rebuildForm: Replacing form in container")
	fb.container.Clear()
	fb.container.AddItem(form, 0, 1, true)
	fb.logger.Info("rebuildForm: Complete")
}

// createAddKeyValueHandler creates the handler for adding key-value pairs
func (fb *KeyValueFormBuilder) createAddKeyValueHandler() func() {
	return func() {
		fb.logger.Info("Add Key-Value button clicked")

		keyField, valueField := fb.findKeyValueFields()
		if keyField == nil || valueField == nil {
			fb.logger.Error("Could not find Key and Value fields")
			return
		}
		fb.logger.Info("Found Key and Value fields")

		// Trim key but preserve whitespace in value (including newlines)
		key := strings.TrimSpace(keyField.GetText())
		value := valueField.GetText()
		fb.logger.Infof("Retrieved key='%s', value length=%d", key, len(value))

		if value == "" {
			fb.logger.Warn("Value cannot be empty")
			return
		}

		// Allow empty key when value is present - use empty string as marker
		if key == "" {
			fb.logger.Info("Key is empty but value is present - will use empty key marker")
		}

		// Add to the pairs map
		fb.keyValuePairs[key] = value
		fb.logger.Infof("Added key-value pair: %s", key)

		// Rebuild the form (no need for QueueUpdateDraw since we're already in UI thread)
		fb.logger.Info("Starting rebuildForm")
		fb.rebuildForm()
		fb.logger.Info("Completed rebuildForm")
		if fb.currentForm != nil {
			fb.app.SetFocus(fb.currentForm)
			fb.logger.Info("Focus set to currentForm")
		}
	}
}

// createSaveHandler creates the handler for saving the form
func (fb *KeyValueFormBuilder) createSaveHandler() func() {
	return func() {
		fb.logger.Info("Save button clicked")

		// Collect all existing key-value fields (for edit mode)
		existingFields := fb.collectAllKeyValueFields()
		fb.logger.Infof("Collected %d existing key-value fields", len(existingFields))

		// Update the internal keyValuePairs map with current values from existing fields
		for key, textArea := range existingFields {
			currentValue := textArea.GetText()
			fb.keyValuePairs[key] = currentValue
			fb.logger.Infof("Updated key '%s' with current value (length=%d)", key, len(currentValue))
		}

		// Auto-add unsaved key-value fields before saving (for new pairs)
		keyField, valueField := fb.findKeyValueFields()
		if keyField != nil && valueField != nil {
			// Trim key but preserve whitespace in value (including newlines)
			key := strings.TrimSpace(keyField.GetText())
			value := valueField.GetText()

			// Allow empty key when value is present
			if value != "" {
				fb.keyValuePairs[key] = value
				fb.logger.Infof("Auto-added new key-value pair before save: key='%s'", key)
			}
		}

		fb.logger.Infof("Final keyValuePairs count: %d", len(fb.keyValuePairs))

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

		// Rebuild the form (no need for QueueUpdateDraw since we're already in UI thread)
		fb.rebuildForm()
		if fb.currentForm != nil {
			fb.app.SetFocus(fb.currentForm)
		}
	}
}

// findKeyValueFields finds the Key and Value input fields in the form
func (fb *KeyValueFormBuilder) findKeyValueFields() (*tview.InputField, *tview.TextArea) {
	fb.logger.Info("findKeyValueFields: Starting search")
	var keyField *tview.InputField
	var valueField *tview.TextArea

	if fb.currentForm == nil {
		fb.logger.Error("findKeyValueFields: currentForm is nil")
		return nil, nil
	}

	formItemCount := fb.currentForm.GetFormItemCount()
	fb.logger.Infof("findKeyValueFields: Form has %d items", formItemCount)

	// Scan backward from buttons to find the new Key and Value fields
	for i := fb.currentForm.GetFormItemCount() - 1; i >= 0; i-- {
		item := fb.currentForm.GetFormItem(i)
		if inputField, ok := item.(*tview.InputField); ok {
			label := inputField.GetLabel()
			fb.logger.Infof("findKeyValueFields: Found InputField at index %d with label='%s'", i, label)
			if label == "Key" && keyField == nil {
				keyField = inputField
				fb.logger.Info("findKeyValueFields: Found Key field")
			}
		} else if textArea, ok := item.(*tview.TextArea); ok {
			label := textArea.GetLabel()
			fb.logger.Infof("findKeyValueFields: Found TextArea at index %d with label='%s'", i, label)
			if label == "Value" && valueField == nil {
				valueField = textArea
				fb.logger.Info("findKeyValueFields: Found Value field")
			}
		}
	}

	fb.logger.Infof("findKeyValueFields: Result - keyField=%v, valueField=%v", keyField != nil, valueField != nil)
	return keyField, valueField
}

// collectAllKeyValueFields collects all key-value fields from the form for edit mode
func (fb *KeyValueFormBuilder) collectAllKeyValueFields() map[string]*tview.TextArea {
	fb.logger.Info("collectAllKeyValueFields: Starting collection")
	keyValueFields := make(map[string]*tview.TextArea)

	if fb.currentForm == nil {
		fb.logger.Error("collectAllKeyValueFields: currentForm is nil")
		return keyValueFields
	}

	formItemCount := fb.currentForm.GetFormItemCount()
	fb.logger.Infof("collectAllKeyValueFields: Form has %d items", formItemCount)

	// Scan all form items to find TextArea fields that represent existing keys
	for i := 0; i < formItemCount; i++ {
		item := fb.currentForm.GetFormItem(i)
		if textArea, ok := item.(*tview.TextArea); ok {
			label := textArea.GetLabel()
			fb.logger.Infof("collectAllKeyValueFields: Found TextArea at index %d with label='%s'", i, label)

			// Skip the "Value" field for new pairs, collect all others (existing keys)
			if label != "Value" {
				keyValueFields[label] = textArea
				fb.logger.Infof("collectAllKeyValueFields: Collected field for key='%s'", label)
			}
		}
	}

	fb.logger.Infof("collectAllKeyValueFields: Collected %d key-value fields", len(keyValueFields))
	return keyValueFields
}
