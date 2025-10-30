package engines

import (
	"testing"

	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnginesFactory_SetupEngine(t *testing.T) {
	factory := NewEnginesFactory(logrus.New())
	assert.NotNil(t, factory)

	t.Run("vault", func(t *testing.T) {
		engine, err := factory.SetupEngine("vault", &config.Profile{
			Engine:     "vault",
			Address:    "http://localhost:8200",
			AuthMethod: "token",
		})
		require.NoError(t, err)
		assert.NotNil(t, engine)
	})

	t.Run("unknown", func(t *testing.T) {
		engine, err := factory.SetupEngine("unknown", &config.Profile{
			Engine:  "unknown",
			Address: "http://localhost:8200",
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unknown engine")
		assert.Nil(t, engine)
	})
}
