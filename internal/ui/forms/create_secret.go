package forms

import (
	"github.com/rivo/tview"
)

// CreateSecretForm creates a form for creating a new secret
func (fm *FormsManager) CreateSecretForm(basePath string, callback func()) tview.Primitive {
	// Initialize path
	secretPath := basePath + "/"

	// Create form builder with create-specific configuration
	builder := NewKeyValueFormBuilder(fm.app, fm.logger, KeyValueFormConfig{
		Title:          "Create New Secret",
		PathLabel:      "Secret Path",
		PathEditable:   true,
		ShowDeleteKeys: false,
		SaveButtonText: "Create",
		OnSave: func(path string, data map[string]string) {
			// Convert to interface map for vault
			secretData := make(map[string]interface{})
			for k, v := range data {
				secretData[k] = v
			}

			// Use handler with callbacks
			fm.secretHandler.CreateSecretWithCallbacks(
				path,
				secretData,
				func() {
					fm.logger.Infof("Created secret: %s with %d key-value pairs", path, len(data))
					if callback != nil {
						callback()
					}
				},
				func(err error) {
					fm.logger.Errorf("Failed to create secret: %v", err)
				},
			)
		},
		OnCancel: callback,
	})

	// Set initial path
	builder.SetPath(secretPath)

	return builder.Build()
}
