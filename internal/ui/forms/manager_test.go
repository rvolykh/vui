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

func TestNewFormsManager(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	assert.NotNil(t, fm)
	assert.Equal(t, cfg, fm.config)
	assert.Equal(t, interactor, fm.interactor)
	assert.Equal(t, logger, fm.logger)
	assert.Equal(t, app, fm.app)
	assert.NotNil(t, fm.secretHandler)
}

func TestFormsManager_GetSecretHandler(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)
	handler := fm.GetSecretHandler()

	assert.NotNil(t, handler)
	assert.Equal(t, fm.secretHandler, handler)
}

func TestFormsManager_GetSecretHandler_NotNil(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	fm := NewFormsManager(cfg, interactor, logger, app)

	// Get handler multiple times should return the same instance
	handler1 := fm.GetSecretHandler()
	handler2 := fm.GetSecretHandler()

	assert.Equal(t, handler1, handler2)
}
