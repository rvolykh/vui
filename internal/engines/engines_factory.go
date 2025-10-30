package engines

import (
	"fmt"

	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/engines/aws"
	"github.com/rvolykh/vui/internal/engines/vault"
	"github.com/sirupsen/logrus"
)

type EnginesFactory struct {
	logger *logrus.Logger
}

func NewEnginesFactory(logger *logrus.Logger) *EnginesFactory {
	return &EnginesFactory{
		logger: logger,
	}
}

func (f *EnginesFactory) SetupEngine(name string, profile *config.Profile) (SecretEngine, error) {
	switch name {
	case "vault":
		return vault.NewVaultClient(f.logger, profile)
	case "aws/secretsmanager":
		return aws.NewAWSSecretsManagerClient(f.logger, profile)
	default:
		return nil, fmt.Errorf("unknown engine: %s", name)
	}
}
