package vault_test

import (
	"testing"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/engines"
	"github.com/rvolykh/vui/internal/engines/vault"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestVaultClient(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	profile := &config.VaultProfile{
		Address: "http://localhost:8200",
	}

	client, err := vault.NewVaultClient(logger, profile)
	require.NoError(t, err)

	require.Implements(t, (*engines.SecretEngine)(nil), client)
}
