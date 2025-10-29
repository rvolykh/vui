package backend

import (
	"testing"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/engines/fake"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInteractor_Implements(t *testing.T) {
	assert.Implements(t, (*Interactor)(nil), &interactor{})
}

func TestInteractor_NewInteractor(t *testing.T) {
	cfg := &config.Config{}
	logger := logrus.New()

	interactor, err := NewInteractor(logger, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, interactor)

	t.Run("Profiles", func(t *testing.T) {
		profiles := interactor.Profiles()
		assert.NotNil(t, profiles)
	})

	t.Run("Secrets nil", func(t *testing.T) {
		secrets, err := interactor.Secrets()
		assert.Error(t, err)
		assert.Nil(t, secrets)
	})

	t.Run("Secrets ok", func(t *testing.T) {
		interactor.profilesInteractor.connectionMgr.AddConnection("test", fake.NewFakeClient())
		interactor.profilesInteractor.SwitchProfile("test")

		secrets, err := interactor.Secrets()
		assert.NoError(t, err)
		assert.NotNil(t, secrets)
	})
}
