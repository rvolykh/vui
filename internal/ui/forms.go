package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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

// showModal shows or hides a modal
func (fm *FormsManager) showModal(primitive tview.Primitive, show bool) {
	if fm.modalHandler != nil {
		fm.modalHandler(primitive, show)
	}
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
				fm.handleCreateSecretWithPairs(form, keyValuePairs, secretPath, callback)
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
	form := tview.NewForm()

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

	// Add form fields for each key-value pair
	form.AddInputField("Secret Path", secretPath, 50, nil, nil)

	// Add fields for existing key-value pairs
	if secret.Data != nil {
		// Sort keys alphabetically
		keys := make([]string, 0, len(secret.Data))
		for key := range secret.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Add fields in sorted order
		for _, key := range keys {
			value := secret.Data[key]
			valueStr := fmt.Sprintf("%v", value)
			form.AddInputField("Key: "+key, valueStr, 50, nil, nil)
		}
	}

	form.AddButton("Add Key-Value", func() {
		// This would add another key-value pair
		fm.logger.Info("Add key-value pair")
	}).
		AddButton("Save", func() {
			fm.handleEditSecret(form, secretPath, callback)
		}).
		AddButton("Cancel", func() {
			if callback != nil {
				callback()
			}
		})

	form.SetBorder(true).
		SetTitle("Edit Secret: " + secret.Name).
		SetTitleAlign(tview.AlignLeft)

	return form
}

// DeleteSecretForm creates a confirmation form for deleting a secret
func (fm *FormsManager) DeleteSecretForm(secretPath string, callback func()) tview.Primitive {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Are you sure you want to delete the secret:\n\n%s\n\nThis action cannot be undone.", secretPath)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				fm.handleDeleteSecret(secretPath, callback)
			} else {
				if callback != nil {
					callback()
				}
			}
		})

	return modal
}

// SearchForm creates a form for searching secrets
func (fm *FormsManager) SearchForm(callback func()) tview.Primitive {
	form := tview.NewForm()

	form.AddInputField("Search Pattern", "", 30, nil, nil).
		AddInputField("Search Path", "", 30, nil, nil).
		AddButton("Search", func() {
			fm.handleSearch(form, callback)
		}).
		AddButton("Advanced", func() {
			fm.showAdvancedSearch(callback)
		}).
		AddButton("Cancel", func() {
			if callback != nil {
				callback()
			}
		})

	form.SetBorder(true).
		SetTitle("Search Secrets").
		SetTitleAlign(tview.AlignLeft)

	return form
}

// AdvancedSearchForm creates an advanced search form
func (fm *FormsManager) AdvancedSearchForm(callback func()) tview.Primitive {
	form := tview.NewForm()

	// Search pattern
	form.AddInputField("Search Pattern", "", 30, nil, nil)

	// Search path
	form.AddInputField("Search Path", "", 30, nil, nil)

	// Search type dropdown
	searchTypes := []string{"Name", "Path", "Key", "Value", "Metadata", "All"}
	form.AddDropDown("Search Type", searchTypes, 0, nil)

	// Max depth
	form.AddInputField("Max Depth", "10", 10, nil, nil)

	// Case sensitive checkbox
	form.AddCheckbox("Case Sensitive", false, nil)

	// Regex checkbox
	form.AddCheckbox("Use Regex", false, nil)

	// Key filter
	form.AddInputField("Key Filter (optional)", "", 30, nil, nil)

	// Value filter
	form.AddInputField("Value Filter (optional)", "", 30, nil, nil)

	form.AddButton("Search", func() {
		fm.handleAdvancedSearch(form, callback)
	}).
		AddButton("Cancel", func() {
			if callback != nil {
				callback()
			}
		})

	form.SetBorder(true).
		SetTitle("Advanced Search").
		SetTitleAlign(tview.AlignLeft)

	return form
}

// handleCreateSecret handles the creation of a new secret
func (fm *FormsManager) handleCreateSecret(form *tview.Form, callback func()) {
	// Get form values
	pathField := form.GetFormItem(0).(*tview.InputField)
	keyField := form.GetFormItem(1).(*tview.InputField)
	valueField := form.GetFormItem(2).(*tview.InputField)

	path := strings.TrimSpace(pathField.GetText())
	key := strings.TrimSpace(keyField.GetText())
	value := strings.TrimSpace(valueField.GetText())

	if path == "" || key == "" || value == "" {
		fm.logger.Error("All fields are required")
		return
	}

	// Create the secret data
	secretData := map[string]interface{}{
		key: value,
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

	fm.logger.Infof("Created secret: %s", path)

	if callback != nil {
		callback()
	}
}

// handleCreateSecretWithPairs handles the creation of a new secret with multiple key-value pairs
func (fm *FormsManager) handleCreateSecretWithPairs(form *tview.Form, keyValuePairs map[string]string, path string, callback func()) {
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

// handleEditSecret handles the editing of an existing secret
func (fm *FormsManager) handleEditSecret(form *tview.Form, secretPath string, callback func()) {
	// Get secrets manager
	secretsManager, err := fm.vaultMgr.GetSecretsManager()
	if err != nil {
		fm.logger.Errorf("Failed to get secrets manager: %v", err)
		return
	}

	// Build secret data from form fields
	secretData := make(map[string]interface{})

	// Skip the first field (path) and process key-value pairs
	for i := 1; i < form.GetFormItemCount()-3; i++ { // -3 for buttons
		field := form.GetFormItem(i).(*tview.InputField)
		label := field.GetLabel()

		if strings.HasPrefix(label, "Key: ") {
			key := strings.TrimPrefix(label, "Key: ")
			value := strings.TrimSpace(field.GetText())
			secretData[key] = value
		}
	}

	// Update the secret
	if err := secretsManager.UpdateSecret(secretPath, secretData); err != nil {
		fm.logger.Errorf("Failed to update secret: %v", err)
		return
	}

	fm.logger.Infof("Updated secret: %s", secretPath)

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

// handleSearch handles the search functionality
func (fm *FormsManager) handleSearch(form *tview.Form, callback func()) {
	patternField := form.GetFormItem(0).(*tview.InputField)
	pathField := form.GetFormItem(1).(*tview.InputField)

	pattern := strings.TrimSpace(patternField.GetText())
	path := strings.TrimSpace(pathField.GetText())

	if pattern == "" {
		fm.logger.Error("Search pattern is required")
		return
	}

	// Get secrets manager
	secretsManager, err := fm.vaultMgr.GetSecretsManager()
	if err != nil {
		fm.logger.Errorf("Failed to get secrets manager: %v", err)
		return
	}

	// Perform search
	results, err := secretsManager.SearchSecrets(pattern, path)
	if err != nil {
		fm.logger.Errorf("Failed to search secrets: %v", err)
		return
	}

	fm.logger.Infof("Search found %d results", len(results))

	if callback != nil {
		callback()
	}
}

// showAdvancedSearch shows the advanced search form
func (fm *FormsManager) showAdvancedSearch(callback func()) {
	advancedForm := fm.AdvancedSearchForm(callback)

	// Create a modal with the advanced search form
	modal := tview.NewModal().
		SetText("").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if callback != nil {
				callback()
			}
		})

	// Replace the modal content with the advanced search form
	modal.SetText("").SetBorder(false)

	// Create a flex layout to contain the form
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(advancedForm, 0, 1, true)

	// Show the advanced search form
	fm.showModal(flex, true)
}

// handleAdvancedSearch handles the advanced search functionality
func (fm *FormsManager) handleAdvancedSearch(form *tview.Form, callback func()) {
	// Get form values
	patternField := form.GetFormItem(0).(*tview.InputField)
	pathField := form.GetFormItem(1).(*tview.InputField)
	searchTypeDropdown := form.GetFormItem(2).(*tview.DropDown)
	maxDepthField := form.GetFormItem(3).(*tview.InputField)
	caseSensitiveCheckbox := form.GetFormItem(4).(*tview.Checkbox)
	regexCheckbox := form.GetFormItem(5).(*tview.Checkbox)
	keyFilterField := form.GetFormItem(6).(*tview.InputField)
	valueFilterField := form.GetFormItem(7).(*tview.InputField)

	pattern := strings.TrimSpace(patternField.GetText())
	path := strings.TrimSpace(pathField.GetText())
	maxDepthStr := strings.TrimSpace(maxDepthField.GetText())
	keyFilter := strings.TrimSpace(keyFilterField.GetText())
	valueFilter := strings.TrimSpace(valueFilterField.GetText())

	if pattern == "" {
		fm.logger.Error("Search pattern is required")
		return
	}

	// Parse max depth
	maxDepth := 10
	if maxDepthStr != "" {
		if d, err := strconv.Atoi(maxDepthStr); err == nil {
			maxDepth = d
		}
	}

	// Get search type
	searchTypeIndex, _ := searchTypeDropdown.GetCurrentOption()
	var searchTypeEnum vault.SearchType
	switch searchTypeIndex {
	case 0: // Name
		searchTypeEnum = vault.SearchType(0) // SearchByName
	case 1: // Path
		searchTypeEnum = vault.SearchType(1) // SearchByPath
	case 2: // Key
		searchTypeEnum = vault.SearchType(2) // SearchByKey
	case 3: // Value
		searchTypeEnum = vault.SearchType(3) // SearchByValue
	case 4: // Metadata
		searchTypeEnum = vault.SearchType(4) // SearchByMetadata
	case 5: // All
		searchTypeEnum = vault.SearchType(5) // SearchAll
	default:
		searchTypeEnum = vault.SearchType(0) // SearchByName
	}

	// Get checkbox values
	caseSensitive := caseSensitiveCheckbox.IsChecked()
	useRegex := regexCheckbox.IsChecked()

	// Get secrets manager
	secretsManager, err := fm.vaultMgr.GetSecretsManager()
	if err != nil {
		fm.logger.Errorf("Failed to get secrets manager: %v", err)
		return
	}

	// Create search options
	options := &vault.AdvancedSearchOptions{
		Pattern:       pattern,
		RootPath:      path,
		SearchType:    searchTypeEnum,
		KeyFilter:     keyFilter,
		ValueFilter:   valueFilter,
		MaxDepth:      maxDepth,
		CaseSensitive: caseSensitive,
		Regex:         useRegex,
	}

	// Perform advanced search
	results, err := secretsManager.AdvancedSearch(options)
	if err != nil {
		fm.logger.Errorf("Failed to perform advanced search: %v", err)
		return
	}

	fm.logger.Infof("Advanced search found %d results", len(results))

	// Show search results
	fm.showSearchResults(results, callback)
}

// showSearchResults displays search results in a modal
func (fm *FormsManager) showSearchResults(results []*vault.SearchResult, callback func()) {
	if len(results) == 0 {
		modal := tview.NewModal().
			SetText("No search results found.").
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if callback != nil {
					callback()
				}
			})
		fm.showModal(modal, false)
		return
	}

	// Create a list to display results
	list := tview.NewList()

	for i, result := range results {
		// Limit display to first 50 results
		if i >= 50 {
			list.AddItem(fmt.Sprintf("... and %d more results", len(results)-50), "", 0, nil)
			break
		}

		// Create display text
		displayText := fmt.Sprintf("%s (Score: %.1f)", result.Node.Name, result.Score)
		secondaryText := fmt.Sprintf("Path: %s | Match: %s", result.Path, result.MatchType)

		// Add item to list
		list.AddItem(displayText, secondaryText, 0, func() {
			// Handle selection - could navigate to the secret
			fm.logger.Infof("Selected search result: %s", result.Path)
		})
	}

	list.SetTitle(fmt.Sprintf("Search Results (%d found)", len(results))).
		SetBorder(true)

	list.SetDoneFunc(func() {
		if callback != nil {
			callback()
		}
	})

	fm.showModal(list, true)
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
		})

	return modal
}
