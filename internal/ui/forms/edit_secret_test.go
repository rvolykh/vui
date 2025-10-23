package forms

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFormsManager_EditSecretForm_WithGetSecretError(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	callback := func() {
		// Callback for testing
	}

	// This will fail because vaultMgr is not properly initialized
	// and GetSecret will return an error
	primitive := fm.EditSecretForm("/secret/test", callback)

	assert.NotNil(t, primitive)
	// Should return an error modal
	modal, ok := primitive.(*tview.Modal)
	assert.True(t, ok)
	assert.NotNil(t, modal)
}

func TestFormsManager_EditSecretForm_WithEmptyPath(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	// Empty path should cause GetSecret to fail
	primitive := fm.EditSecretForm("", func() {})

	assert.NotNil(t, primitive)
	// Should return an error modal
	_, ok := primitive.(*tview.Modal)
	assert.True(t, ok)
}

func TestFormsManager_EditSecretForm_WithNilCallback(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	// Should not panic with nil callback
	primitive := fm.EditSecretForm("/secret/test", nil)

	assert.NotNil(t, primitive)
}

func TestFormsManager_EditSecretForm_MultipleInvocations(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	// Create multiple edit forms
	form1 := fm.EditSecretForm("/secret/path1", func() {})
	form2 := fm.EditSecretForm("/secret/path2", func() {})

	assert.NotNil(t, form1)
	assert.NotNil(t, form2)
}

func TestFormsManager_EditSecretForm_WithLongPath(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	longPath := "/secret/very/long/path/to/secret/that/needs/to/be/edited"
	primitive := fm.EditSecretForm(longPath, func() {})

	assert.NotNil(t, primitive)
}

func TestFormsManager_EditSecretForm_WithSpecialCharactersInPath(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	specialPath := "/secret/test-secret_123/item.name"
	primitive := fm.EditSecretForm(specialPath, func() {})

	assert.NotNil(t, primitive)
}

func TestFormsManager_EditSecretForm_ReturnsErrorModalOnFailure(t *testing.T) {
	cfg := &config.Config{}
	vaultMgr := &vault.Manager{}
	logger := logrus.New()
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, vaultMgr, logger, app)

	// Since vault is not initialized, GetSecret will fail
	primitive := fm.EditSecretForm("/secret/test", func() {})

	// Should return a modal (error modal)
	modal, ok := primitive.(*tview.Modal)
	assert.True(t, ok, "Expected error modal to be returned")
	assert.NotNil(t, modal)
}
