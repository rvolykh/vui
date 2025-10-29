package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_Run(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	assert.NotNil(t, app)

	// Set TERM to empty string to force tcell to fail immediately
	require.NoError(t, os.Setenv("TERM", ""))

	err = app.Run()
	assert.Error(t, err)
}
