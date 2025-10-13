package forms

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/ui/common"
)

// EditSecretForm creates a form for editing an existing secret
func (fm *FormsManager) EditSecretForm(secretPath string, callback func()) tview.Primitive {
	// Get the current secret
	secret, err := fm.secretHandler.GetSecret(secretPath)
	if err != nil {
		fm.logger.Errorf("Failed to get secret: %v", err)
		// Return error modal
		return common.ErrorModal(fmt.Sprintf("Failed to get secret: %v", err), callback)
	}

	// Store key-value pairs that will be edited
	initialData := make(map[string]string)

	// Initialize with existing data
	if secret.Data != nil {
		for key, value := range secret.Data {
			initialData[key] = fmt.Sprintf("%v", value)
		}
	}

	// Create form builder with edit-specific configuration
	builder := NewKeyValueFormBuilder(fm.app, fm.logger, KeyValueFormConfig{
		Title:          fmt.Sprintf("Edit Secret: %s", secret.Name),
		PathLabel:      "Secret Path",
		PathEditable:   false,
		InitialData:    initialData,
		ShowDeleteKeys: true,
		SaveButtonText: "Save",
		OnSave: func(path string, data map[string]string) {
			// Convert to interface map for vault
			secretData := make(map[string]interface{})
			for k, v := range data {
				secretData[k] = v
			}

			// Use handler with callbacks
			fm.secretHandler.UpdateSecretWithCallbacks(
				path,
				secretData,
				func() {
					fm.logger.Infof("Updated secret: %s with %d key-value pairs", path, len(data))
					if callback != nil {
						callback()
					}
				},
				func(err error) {
					fm.logger.Errorf("Failed to update secret: %v", err)
				},
			)
		},
		OnCancel: callback,
	})

	// Set the path (read-only for edit)
	builder.SetPath(secretPath)

	return builder.Build()
}
