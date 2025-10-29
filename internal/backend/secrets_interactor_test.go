package backend

import (
	"errors"
	"testing"

	"github.com/rvolykh/vui/internal/engines/fake"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSecretsInteractor_Implements(t *testing.T) {
	assert.Implements(t, (*SecretsInteractor)(nil), &secretsInteractor{})
}

func TestSecretsInteractor_BuildTree(t *testing.T) {
	t.Run("secret", func(t *testing.T) {
		client := fake.NewFakeClient()
		client.RespListSecrets = []*models.SecretNode{{Name: "test", Path: "test", IsSecret: true}}
		interactor := newSecretsInteractor(logrus.New(), "test", client)

		tree, err := interactor.BuildTree("test", 1)

		assert.NoError(t, err)
		assert.NotNil(t, tree)
	})

	t.Run("folder", func(t *testing.T) {
		client := fake.NewFakeClient()
		client.RespListSecrets = []*models.SecretNode{{Name: "test", Path: "test", IsSecret: false}}
		interactor := newSecretsInteractor(logrus.New(), "test", client)

		tree, err := interactor.BuildTree("test", 1)

		assert.NoError(t, err)
		assert.NotNil(t, tree)
	})

	t.Run("no secrets", func(t *testing.T) {
		client := fake.NewFakeClient()
		client.RespListSecrets = []*models.SecretNode{}
		interactor := newSecretsInteractor(logrus.New(), "test", client)

		tree, err := interactor.BuildTree("", 1)

		assert.NoError(t, err)
		assert.NotNil(t, tree)
	})

	t.Run("list failure", func(t *testing.T) {
		client := fake.NewFakeClient()
		client.RespErr = errors.New("test error")
		interactor := newSecretsInteractor(logrus.New(), "test", client)

		tree, err := interactor.BuildTree("test", 1)

		assert.Error(t, err)
		assert.Nil(t, tree)
	})
}
