package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// FormsManager manages input forms for the application
type FormsManager struct {
	config       *config.Config
	vaultMgr     *vault.Manager
	logger       *logrus.Logger
	modalHandler func(tview.Primitive, bool)
	app          *tview.Application
}

// NewFormsManager creates a new forms manager
func NewFormsManager(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger, app *tview.Application) *FormsManager {
	return &FormsManager{
		config:   config,
		vaultMgr: vaultMgr,
		logger:   logger,
		app:      app,
	}
}

// SetModalHandler sets the modal handler for showing modals
func (fm *FormsManager) SetModalHandler(handler func(tview.Primitive, bool)) {
	fm.modalHandler = handler
}

// CreateSecretForm creates a form for creating a new secret
func (fm *FormsManager) CreateSecretForm(basePath string, callback func()) tview.Primitive {
	// Store key-value pairs that have been added
	keyValuePairs := make(map[string]string)

	// Store the secret path (persists across rebuilds)
	secretPath := basePath + "/"

	// Create a flex container to hold the form and allow dynamic updates
	container := tview.NewFlex().SetDirection(tview.FlexRow)

	// Store reference to the current form for focus management
	var currentForm *tview.Form

	// Function to rebuild the form with current key-value pairs
	var rebuildForm func()
	rebuildForm = func() {
		form := tview.NewForm()
		currentForm = form

		// Secret path field - make it read-only if we have pairs added
		pathField := tview.NewInputField().
			SetLabel("Secret Path").
			SetText(secretPath).
			SetFieldWidth(50)

		// Make the path field read-only if pairs have been added
		if len(keyValuePairs) > 0 {
			pathField.SetDisabled(true)
		} else {
			// Allow editing and update secretPath when changed
			pathField.SetChangedFunc(func(text string) {
				secretPath = text
			})
		}

		form.AddFormItem(pathField)

		// Show existing key-value pairs (read-only display)
		if len(keyValuePairs) > 0 {
			pairText := fmt.Sprintf("Added Pairs (%d):", len(keyValuePairs))
			form.AddTextView(pairText, "", 0, 1, false, false)
			for k, v := range keyValuePairs {
				displayValue := v
				if len(displayValue) > 40 {
					displayValue = displayValue[:37] + "..."
				}
				form.AddTextView("  • "+k, displayValue, 0, 1, false, false)
			}
		}

		form.AddTextView("New Pair:", "", 0, 1, false, false)

		// New key-value input fields
		form.AddInputField("Key", "", 30, nil, nil).
			AddInputField("Value", "", 50, nil, nil)

		// Buttons
		form.AddButton("Add Key-Value", func() {
			// Find the Key and Value input fields by scanning backward from buttons
			var keyField, valueField *tview.InputField
			for i := form.GetFormItemCount() - 1; i >= 0; i-- {
				item := form.GetFormItem(i)
				if inputField, ok := item.(*tview.InputField); ok {
					label := inputField.GetLabel()
					if label == "Value" && valueField == nil {
						valueField = inputField
					} else if label == "Key" && keyField == nil {
						keyField = inputField
					}
				}
			}

			if keyField == nil || valueField == nil {
				fm.logger.Error("Could not find Key and Value fields")
				return
			}

			key := strings.TrimSpace(keyField.GetText())
			value := strings.TrimSpace(valueField.GetText())

			if key == "" {
				fm.logger.Warn("Key cannot be empty")
				return
			}

			if value == "" {
				fm.logger.Warn("Value cannot be empty")
				return
			}

			// Add to the pairs map
			keyValuePairs[key] = value
			fm.logger.Infof("Added key-value pair: %s", key)

			// Rebuild the form in a goroutine to avoid blocking
			go func() {
				if fm.app != nil {
					fm.app.QueueUpdateDraw(func() {
						rebuildForm()
						// Set focus back to the form after rebuild
						if currentForm != nil {
							fm.app.SetFocus(currentForm)
						}
					})
				}
			}()
		}).
			AddButton("Create", func() {
				// Before creating, check if there are unsaved key-value fields
				var keyField, valueField *tview.InputField
				for i := form.GetFormItemCount() - 1; i >= 0; i-- {
					item := form.GetFormItem(i)
					if inputField, ok := item.(*tview.InputField); ok {
						label := inputField.GetLabel()
						if label == "Value" && valueField == nil {
							valueField = inputField
						} else if label == "Key" && keyField == nil {
							keyField = inputField
						}
					}
				}

				// If there are unsaved key-value fields, add them first
				if keyField != nil && valueField != nil {
					key := strings.TrimSpace(keyField.GetText())
					value := strings.TrimSpace(valueField.GetText())

					if key != "" && value != "" {
						keyValuePairs[key] = value
						fm.logger.Infof("Auto-added key-value pair before create: %s", key)
					}
				}

				// Now create the secret with all pairs
				fm.handleCreateSecretWithPairs(keyValuePairs, secretPath, callback)
			}).
			AddButton("Cancel", func() {
				if callback != nil {
					callback()
				}
			})

		form.SetBorder(true).
			SetTitle(fmt.Sprintf("Create New Secret%s", func() string {
				if len(keyValuePairs) > 0 {
					return fmt.Sprintf(" [%d pairs added]", len(keyValuePairs))
				}
				return ""
			}())).
			SetTitleAlign(tview.AlignLeft)

		// Set form field colors for a clean, neutral look
		form.SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
			SetFieldTextColor(tcell.ColorWhite).
			SetLabelColor(tcell.ColorLightGray)

		// Set button colors - inactive buttons have dark gray background with white text
		// active/focused buttons have cyan background with black text for clear visual distinction
		form.SetButtonBackgroundColor(tcell.ColorDarkGray).
			SetButtonTextColor(tcell.ColorWhite).
			SetButtonActivatedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorBlack))

		// Replace the form in the container
		container.Clear()
		container.AddItem(form, 0, 1, true)
	}

	// Initial form build
	rebuildForm()

	return container
}

// EditSecretForm creates a form for editing an existing secret
func (fm *FormsManager) EditSecretForm(secretPath string, callback func()) tview.Primitive {
	// Get the current secret
	secretsManager, err := fm.vaultMgr.GetSecretsManager()
	if err != nil {
		fm.logger.Errorf("Failed to get secrets manager: %v", err)
		return fm.createErrorForm("Failed to get secrets manager", callback)
	}

	secret, err := secretsManager.GetSecret(secretPath)
	if err != nil {
		fm.logger.Errorf("Failed to get secret: %v", err)
		return fm.createErrorForm("Failed to get secret", callback)
	}

	// Store key-value pairs that will be edited
	keyValuePairs := make(map[string]string)

	// Initialize with existing data
	if secret.Data != nil {
		for key, value := range secret.Data {
			keyValuePairs[key] = fmt.Sprintf("%v", value)
		}
	}

	// Create a flex container to hold the form and allow dynamic updates
	container := tview.NewFlex().SetDirection(tview.FlexRow)

	// Store reference to the current form for focus management
	var currentForm *tview.Form

	// Function to rebuild the form with current key-value pairs
	var rebuildForm func()
	rebuildForm = func() {
		form := tview.NewForm()
		currentForm = form

		// Secret path field (read-only)
		pathField := tview.NewInputField().
			SetLabel("Secret Path").
			SetText(secretPath).
			SetFieldWidth(50)
		pathField.SetDisabled(true)
		form.AddFormItem(pathField)

		// Show existing key-value pairs (read-only display)
		if len(keyValuePairs) > 0 {
			pairText := fmt.Sprintf("Existing Pairs (%d):", len(keyValuePairs))
			form.AddTextView(pairText, "", 0, 1, false, false)

			// Sort keys alphabetically for consistent display
			keys := make([]string, 0, len(keyValuePairs))
			for k := range keyValuePairs {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				v := keyValuePairs[k]
				displayValue := v
				if len(displayValue) > 40 {
					displayValue = displayValue[:37] + "..."
				}
				form.AddInputField(k, displayValue, 50, nil, nil)
				form.AddButton("Delete "+k, func() {
					keyToDelete := k // Capture the current key value
					delete(keyValuePairs, keyToDelete)
					fm.logger.Infof("Deleted key-value pair: %s", keyToDelete)

					// Rebuild the form in a goroutine to avoid blocking
					go func() {
						if fm.app != nil {
							fm.app.QueueUpdateDraw(func() {
								rebuildForm()
								// Set focus back to the form after rebuild
								if currentForm != nil {
									fm.app.SetFocus(currentForm)
								}
							})
						}
					}()
				})
			}
		}

		form.AddTextView("New Pair:", "", 0, 1, false, false)

		// New key-value input fields
		form.AddInputField("Key", "", 30, nil, nil).
			AddInputField("Value", "", 50, nil, nil)

		// Buttons
		form.AddButton("Add Key-Value", func() {
			// Find the Key and Value input fields by scanning backward from buttons
			var keyField, valueField *tview.InputField
			for i := form.GetFormItemCount() - 1; i >= 0; i-- {
				item := form.GetFormItem(i)
				if inputField, ok := item.(*tview.InputField); ok {
					label := inputField.GetLabel()
					if label == "Value" && valueField == nil {
						valueField = inputField
					} else if label == "Key" && keyField == nil {
						keyField = inputField
					}
				}
			}

			if keyField == nil || valueField == nil {
				fm.logger.Error("Could not find Key and Value fields")
				return
			}

			key := strings.TrimSpace(keyField.GetText())
			value := strings.TrimSpace(valueField.GetText())

			if key == "" {
				fm.logger.Warn("Key cannot be empty")
				return
			}

			if value == "" {
				fm.logger.Warn("Value cannot be empty")
				return
			}

			// Add to the pairs map
			keyValuePairs[key] = value
			fm.logger.Infof("Added key-value pair: %s", key)

			// Rebuild the form in a goroutine to avoid blocking
			go func() {
				if fm.app != nil {
					fm.app.QueueUpdateDraw(func() {
						rebuildForm()
						// Set focus back to the form after rebuild
						if currentForm != nil {
							fm.app.SetFocus(currentForm)
						}
					})
				}
			}()
		}).
			AddButton("Save", func() {
				// Before saving, check if there are unsaved key-value fields
				var keyField, valueField *tview.InputField
				for i := form.GetFormItemCount() - 1; i >= 0; i-- {
					item := form.GetFormItem(i)
					if inputField, ok := item.(*tview.InputField); ok {
						label := inputField.GetLabel()
						if label == "Value" && valueField == nil {
							valueField = inputField
						} else if label == "Key" && keyField == nil {
							keyField = inputField
						}
					}
				}

				// If there are unsaved key-value fields, add them first
				if keyField != nil && valueField != nil {
					key := strings.TrimSpace(keyField.GetText())
					value := strings.TrimSpace(valueField.GetText())

					if key != "" && value != "" {
						keyValuePairs[key] = value
						fm.logger.Infof("Auto-added key-value pair before save: %s", key)
					}
				}

				// Now save the secret with all pairs
				fm.handleEditSecretWithPairs(keyValuePairs, secretPath, callback)
			}).
			AddButton("Cancel", func() {
				if callback != nil {
					callback()
				}
			})

		form.SetBorder(true).
			SetTitle(fmt.Sprintf("Edit Secret: %s%s", secret.Name, func() string {
				if len(keyValuePairs) > 0 {
					return fmt.Sprintf(" [%d pairs]", len(keyValuePairs))
				}
				return ""
			}())).
			SetTitleAlign(tview.AlignLeft)

		// Set form field colors for a clean, neutral look
		form.SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
			SetFieldTextColor(tcell.ColorWhite).
			SetLabelColor(tcell.ColorLightGray)

		// Set button colors for better visibility
		form.SetButtonBackgroundColor(tcell.ColorDarkGray).
			SetButtonTextColor(tcell.ColorWhite).
			SetButtonActivatedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorBlack))

		// Replace the form in the container
		container.Clear()
		container.AddItem(form, 0, 1, true)
	}

	// Initial form build
	rebuildForm()

	return container
}

// DeleteSecretForm creates a confirmation form for deleting a secret
func (fm *FormsManager) DeleteSecretForm(secretPath string, callback func()) tview.Primitive {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Are you sure you want to delete the secret:\n\n[::b]%s[::-]\n\nThis action cannot be undone.", secretPath)).
		AddButtons([]string{"Cancel", "Delete"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				fm.handleDeleteSecret(secretPath, callback)
			} else {
				if callback != nil {
					callback()
				}
			}
		}).
		SetButtonBackgroundColor(tcell.ColorDarkGray).
		SetButtonTextColor(tcell.ColorWhite).
		SetButtonActivatedStyle(
			tcell.StyleDefault.
				Background(tcell.ColorRed).
				Foreground(tcell.ColorWhite).
				Bold(true).
				// Set a border style by using more visible attributes.
				// Note: tcell/tview buttons themselves do not support box borders,
				// so we use Bold and Reverse to simulate a "bordered" effect.
				Reverse(true),
		)

	return modal
}

// handleCreateSecretWithPairs handles the creation of a new secret with multiple key-value pairs
func (fm *FormsManager) handleCreateSecretWithPairs(keyValuePairs map[string]string, path string, callback func()) {
	if path == "" {
		fm.logger.Error("Secret path is required")
		return
	}

	// Check if we have at least one key-value pair
	if len(keyValuePairs) == 0 {
		fm.logger.Error("At least one key-value pair is required")
		return
	}

	// Convert to the format expected by vault
	secretData := make(map[string]interface{})
	for k, v := range keyValuePairs {
		secretData[k] = v
	}

	// Get secrets manager
	secretsManager, err := fm.vaultMgr.GetSecretsManager()
	if err != nil {
		fm.logger.Errorf("Failed to get secrets manager: %v", err)
		return
	}

	// Create the secret
	if err := secretsManager.CreateSecret(path, secretData); err != nil {
		fm.logger.Errorf("Failed to create secret: %v", err)
		return
	}

	fm.logger.Infof("Created secret: %s with %d key-value pairs", path, len(keyValuePairs))

	if callback != nil {
		callback()
	}
}

// handleEditSecretWithPairs handles the editing of an existing secret with multiple key-value pairs
func (fm *FormsManager) handleEditSecretWithPairs(keyValuePairs map[string]string, secretPath string, callback func()) {
	if secretPath == "" {
		fm.logger.Error("Secret path is required")
		return
	}

	// Check if we have at least one key-value pair
	if len(keyValuePairs) == 0 {
		fm.logger.Error("At least one key-value pair is required")
		return
	}

	// Convert to the format expected by vault
	secretData := make(map[string]interface{})
	for k, v := range keyValuePairs {
		secretData[k] = v
	}

	// Get secrets manager
	secretsManager, err := fm.vaultMgr.GetSecretsManager()
	if err != nil {
		fm.logger.Errorf("Failed to get secrets manager: %v", err)
		return
	}

	// Update the secret
	if err := secretsManager.UpdateSecret(secretPath, secretData); err != nil {
		fm.logger.Errorf("Failed to update secret: %v", err)
		return
	}

	fm.logger.Infof("Updated secret: %s with %d key-value pairs", secretPath, len(keyValuePairs))

	if callback != nil {
		callback()
	}
}

// handleDeleteSecret handles the deletion of a secret
func (fm *FormsManager) handleDeleteSecret(secretPath string, callback func()) {
	// Get secrets manager
	secretsManager, err := fm.vaultMgr.GetSecretsManager()
	if err != nil {
		fm.logger.Errorf("Failed to get secrets manager: %v", err)
		return
	}

	// Delete the secret
	if err := secretsManager.DeleteSecret(secretPath); err != nil {
		fm.logger.Errorf("Failed to delete secret: %v", err)
		return
	}

	fm.logger.Infof("Deleted secret: %s", secretPath)

	if callback != nil {
		callback()
	}
}

// createErrorForm creates a simple error form
func (fm *FormsManager) createErrorForm(message string, callback func()) tview.Primitive {
	modal := tview.NewModal().
		SetText("Error: " + message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if callback != nil {
				callback()
			}
		}).
		SetButtonBackgroundColor(tcell.ColorDarkGray).
		SetButtonTextColor(tcell.ColorWhite).
		SetButtonActivatedStyle(tcell.StyleDefault.Background(tcell.ColorDarkCyan).Foreground(tcell.ColorBlack))

	return modal
}
