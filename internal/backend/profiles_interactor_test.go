package backend

import (
	"testing"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/engines/fake"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfilesInteractor_Implements(t *testing.T) {
	assert.Implements(t, (*ProfileInteractor)(nil), &profileInteractor{})
}

func TestProfilesInteractor_newProfileInteractor(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()

	pi, err := newProfileInteractor(logger, cfg)
	require.NoError(t, err)

	assert.NotNil(t, pi)
}

func TestProfilesInteractor_SwitchProfile(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{}

	t.Run("found", func(t *testing.T) {
		pi := profileInteractor{
			logger:        logger,
			config:        cfg,
			connectionMgr: NewConnectionManager(logger),
		}
		pi.connectionMgr.AddConnection("test", fake.NewFakeClient())

		err := pi.SwitchProfile("test")

		assert.NoError(t, err)
		assert.Equal(t, "test", pi.currentProfile)
		assert.NotNil(t, pi.secretsInteractor)
		assert.NotNil(t, pi.connectionMgr)
		assert.NotNil(t, pi.logger)
		assert.NotNil(t, pi.config)
	})

	t.Run("not found", func(t *testing.T) {
		pi := profileInteractor{
			logger:        logger,
			config:        cfg,
			connectionMgr: NewConnectionManager(logger),
		}

		err := pi.SwitchProfile("test")

		assert.Error(t, err)
		assert.Equal(t, "", pi.currentProfile)
		assert.Nil(t, pi.secretsInteractor)
		assert.NotNil(t, pi.connectionMgr)
		assert.NotNil(t, pi.logger)
		assert.NotNil(t, pi.config)
	})
}

func TestProfilesInteractor_ListConnections(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{}

	pi := profileInteractor{
		logger:        logger,
		config:        cfg,
		connectionMgr: NewConnectionManager(logger),
	}
	pi.connectionMgr.AddConnection("test", fake.NewFakeClient())

	have := pi.ListConnections()
	assert.Len(t, have, 1)
	assert.Equal(t, "test", have[0])
}

func TestProfilesInteractor_RefreshConnection(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{}

	pi := profileInteractor{
		logger:        logger,
		config:        cfg,
		connectionMgr: NewConnectionManager(logger),
	}
	client := fake.NewFakeClient()
	client.RespGetStatus = models.ConnectionStatus{
		Status: models.StatusConnected,
	}
	pi.connectionMgr.AddConnection("test", client)

	pi.RefreshConnection("test")

	status, err := pi.GetConnectionStatus("test")
	assert.NoError(t, err)
	assert.Equal(t, models.StatusConnected, status.Status)
}

func TestProfilesInteractor_ResetConnections(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{}

	pi := profileInteractor{
		logger:        logger,
		config:        cfg,
		connectionMgr: NewConnectionManager(logger),
	}
	pi.connectionMgr.AddConnection("test", fake.NewFakeClient())

	pi.ResetConnections()

	status, err := pi.GetConnectionStatus("test")
	assert.NoError(t, err)
	assert.Equal(t, models.StatusConnecting, status.Status)
}

func TestProfilesInteractor_ReloadConfiguration(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{}

	pi := profileInteractor{
		logger:        logger,
		config:        cfg,
		connectionMgr: NewConnectionManager(logger),
	}
	pi.connectionMgr.AddConnection("test", fake.NewFakeClient())

	err := pi.ReloadConfiguration()
	assert.NoError(t, err)

	have := pi.ListConnections()
	assert.Len(t, have, 1)
	assert.Equal(t, "local", have[0])
}

func TestProfilesInteractor_GetCurrentProfile(t *testing.T) {
	logger := logrus.New()
	cfg := &config.Config{}

	pi := profileInteractor{
		logger:        logger,
		config:        cfg,
		connectionMgr: NewConnectionManager(logger),
	}
	pi.connectionMgr.AddConnection("test", fake.NewFakeClient())
	pi.currentProfile = "test"

	have := pi.GetCurrentProfile()
	assert.Equal(t, "test", have)
}
