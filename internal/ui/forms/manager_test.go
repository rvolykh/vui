package forms

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewFormsManager(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	assert.NotNil(t, fm)
	assert.Equal(t, cfg, fm.config)
	assert.Equal(t, vaultMgr, fm.vaultMgr)
	assert.Equal(t, logger, fm.logger)
	assert.Equal(t, app, fm.app)
	assert.NotNil(t, fm.secretHandler)
}

func TestFormsManager_GetSecretHandler(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)
	handler := fm.GetSecretHandler()

	assert.NotNil(t, handler)
	assert.Equal(t, fm.secretHandler, handler)
}

func TestFormsManager_GetSecretHandler_NotNil(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	// Get handler multiple times should return the same instance
	handler1 := fm.GetSecretHandler()
	handler2 := fm.GetSecretHandler()

	assert.Equal(t, handler1, handler2)
}

func TestFormsManager_WithNilLogger(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	app := tview.NewApplication()

	// Should not panic with nil logger
	fm := NewFormsManager(cfg, vaultMgr, nil, app)

	assert.NotNil(t, fm)
	assert.Nil(t, fm.logger)
}

func TestFormsManager_WithNilApp(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()

	// Should not panic with nil app
	fm := NewFormsManager(cfg, vaultMgr, logger, nil)

	assert.NotNil(t, fm)
	assert.Nil(t, fm.app)
}

func TestFormsManager_WithNilConfig(t *testing.T) {
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	// Should not panic with nil config
	fm := NewFormsManager(nil, vaultMgr, logger, app)

	assert.NotNil(t, fm)
	assert.Nil(t, fm.config)
}

func TestFormsManager_WithNilVaultManager(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	app := tview.NewApplication()

	// Should not panic with nil vault manager
	fm := NewFormsManager(cfg, nil, logger, app)

	assert.NotNil(t, fm)
	assert.Nil(t, fm.vaultMgr)
}
