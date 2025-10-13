package forms

import (
	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/ui/handlers"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
)

// FormsManager manages input forms for the application
type FormsManager struct {
	config        *config.Config
	vaultMgr      *vault.Manager
	logger        *logrus.Logger
	app           *tview.Application
	secretHandler *handlers.SecretHandler
}

// NewFormsManager creates a new forms manager
func NewFormsManager(config *config.Config, vaultMgr *vault.Manager, logger *logrus.Logger, app *tview.Application) *FormsManager {
	return &FormsManager{
		config:        config,
		vaultMgr:      vaultMgr,
		logger:        logger,
		app:           app,
		secretHandler: handlers.NewSecretHandler(vaultMgr),
	}
}

// GetSecretHandler returns the secret handler for direct use
func (fm *FormsManager) GetSecretHandler() *handlers.SecretHandler {
	return fm.secretHandler
}
