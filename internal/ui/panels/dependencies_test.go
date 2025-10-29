package panels

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/rvolykh/vui/internal/backend"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// Mock dialog service for testing
type mockDialogService struct{}

func (m *mockDialogService) ShowInfo(title, message string, callback func()) {}
func (m *mockDialogService) ShowError(message string, callback func())       {}

type Fixtures struct {
	cfg        *config.Config
	logger     *logrus.Logger
	interactor backend.Interactor
	app        *tview.Application
}

func WithFixtures(t *testing.T) Fixtures {
	cfg := &config.Config{}
	logger := logrus.New()
	interactor, err := backend.NewInteractor(logger, cfg)
	require.NoError(t, err)
	app := tview.NewApplication()

	return Fixtures{
		cfg:        cfg,
		logger:     logger,
		interactor: interactor,
		app:        app,
	}
}
