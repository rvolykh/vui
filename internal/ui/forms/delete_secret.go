package forms

import (
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/ui/common"
)

// DeleteSecretForm creates a confirmation form for deleting a secret
func (fm *FormsManager) DeleteSecretForm(secretPath string, callback func()) tview.Primitive {
	return common.DeleteConfirmationModal(
		secretPath,
		func() {
			// Use handler with callbacks
			fm.secretHandler.DeleteSecretWithCallbacks(
				secretPath,
				func() {
					fm.logger.Infof("Deleted secret: %s", secretPath)
					if callback != nil {
						callback()
					}
				},
				func(err error) {
					fm.logger.Errorf("Failed to delete secret: %v", err)
				},
			)
		},
		func() {
			if callback != nil {
				callback()
			}
		},
	)
}
