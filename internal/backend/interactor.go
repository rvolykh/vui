package backend

import (
	"fmt"

	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
)

type Interactor interface {
	Secrets() (SecretsInteractor, error)
	Profiles() ProfileInteractor
}

type interactor struct {
	profilesInteractor *profileInteractor
}

func NewInteractor(logger *logrus.Logger, cfg *config.Config) (*interactor, error) {
	profilesInteractor, err := newProfileInteractor(logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile interactor: %w", err)
	}

	return &interactor{
		profilesInteractor: profilesInteractor,
	}, nil
}

func (i *interactor) Secrets() (SecretsInteractor, error) {
	if i.profilesInteractor.secretsInteractor == nil {
		return nil, fmt.Errorf("secrets interactor not found")
	}
	return i.profilesInteractor.secretsInteractor, nil
}

func (i *interactor) Profiles() ProfileInteractor {
	return i.profilesInteractor
}
