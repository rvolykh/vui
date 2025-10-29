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

func TestFormsManager_CreateSecretForm(t *testing.T) {
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

	primitive := fm.CreateSecretForm("/secret/base", callback)

	assert.NotNil(t, primitive)
	assert.False(t, callbackCalled) // Callback not called yet
}

func TestFormsManager_CreateSecretForm_WithEmptyBasePath(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	primitive := fm.CreateSecretForm("", func() {})

	assert.NotNil(t, primitive)
}

func TestFormsManager_CreateSecretForm_WithNilCallback(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	// Should not panic with nil callback
	primitive := fm.CreateSecretForm("/secret/base", nil)

	assert.NotNil(t, primitive)
}

func TestFormsManager_CreateSecretForm_PathInitialization(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	basePath := "/secret/myapp"
	primitive := fm.CreateSecretForm(basePath, func() {})

	assert.NotNil(t, primitive)
	// The form should be a Flex container
	_, ok := primitive.(*tview.Flex)
	assert.True(t, ok)
}

func TestFormsManager_CreateSecretForm_ReturnsFlexContainer(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	primitive := fm.CreateSecretForm("/secret/test", func() {})

	// Should return a Flex container (from builder.Build())
	flex, ok := primitive.(*tview.Flex)
	assert.True(t, ok)
	assert.NotNil(t, flex)
}

func TestFormsManager_CreateSecretForm_WithTrailingSlash(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	primitive := fm.CreateSecretForm("/secret/base/", func() {})

	assert.NotNil(t, primitive)
}

func TestFormsManager_CreateSecretForm_WithoutTrailingSlash(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	primitive := fm.CreateSecretForm("/secret/base", func() {})

	assert.NotNil(t, primitive)
}

func TestFormsManager_CreateSecretForm_MultipleInvocations(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	// Create multiple forms
	form1 := fm.CreateSecretForm("/secret/path1", func() {})
	form2 := fm.CreateSecretForm("/secret/path2", func() {})

	assert.NotNil(t, form1)
	assert.NotNil(t, form2)
	assert.NotEqual(t, form1, form2)
}
