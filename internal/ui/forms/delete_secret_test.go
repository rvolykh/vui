package forms

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/backend"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormsManager_DeleteSecretForm(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)

	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	callbackCalled := false
	callback := func() {
		callbackCalled = true
	}

	primitive := fm.DeleteSecretForm("/secret/test", callback)

	assert.NotNil(t, primitive)
	assert.False(t, callbackCalled) // Callback not called yet
}

func TestFormsManager_DeleteSecretForm_ReturnsModal(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	primitive := fm.DeleteSecretForm("/secret/test", func() {})

	// Should return a Modal (from DeleteConfirmationModal)
	modal, ok := primitive.(*tview.Modal)
	assert.True(t, ok)
	assert.NotNil(t, modal)
}

func TestFormsManager_DeleteSecretForm_WithNilCallback(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	// Should not panic with nil callback
	primitive := fm.DeleteSecretForm("/secret/test", nil)

	assert.NotNil(t, primitive)
}

func TestFormsManager_DeleteSecretForm_WithEmptyPath(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)

	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	primitive := fm.DeleteSecretForm("", func() {})

	assert.NotNil(t, primitive)
}

func TestFormsManager_DeleteSecretForm_WithLongPath(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	longPath := "/secret/very/long/path/to/secret/that/needs/to/be/deleted"
	primitive := fm.DeleteSecretForm(longPath, func() {})

	assert.NotNil(t, primitive)
}

func TestFormsManager_DeleteSecretForm_MultipleInvocations(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	// Create multiple delete forms
	form1 := fm.DeleteSecretForm("/secret/path1", func() {})
	form2 := fm.DeleteSecretForm("/secret/path2", func() {})

	assert.NotNil(t, form1)
	assert.NotNil(t, form2)
	assert.NotEqual(t, form1, form2)
}

func TestFormsManager_DeleteSecretForm_WithSpecialCharactersInPath(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	specialPath := "/secret/test-secret_123/item.name"
	primitive := fm.DeleteSecretForm(specialPath, func() {})

	assert.NotNil(t, primitive)
}
