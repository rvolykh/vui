package aws

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSecretsManager is a mock implementation of secretsmanageriface.SecretsManagerAPI
type mockSecretsManager struct {
	secretsmanageriface.SecretsManagerAPI
	listSecretsOutput    *secretsmanager.ListSecretsOutput
	listSecretsError     error
	getSecretValueOutput *secretsmanager.GetSecretValueOutput
	getSecretValueError  error
	describeSecretOutput *secretsmanager.DescribeSecretOutput
	describeSecretError  error
	createSecretOutput   *secretsmanager.CreateSecretOutput
	createSecretError    error
	updateSecretOutput   *secretsmanager.UpdateSecretOutput
	updateSecretError    error
	deleteSecretOutput   *secretsmanager.DeleteSecretOutput
	deleteSecretError    error
}

func (m *mockSecretsManager) ListSecretsPagesWithContext(ctx aws.Context, input *secretsmanager.ListSecretsInput, fn func(*secretsmanager.ListSecretsOutput, bool) bool, opts ...request.Option) error {
	if m.listSecretsError != nil {
		return m.listSecretsError
	}
	if m.listSecretsOutput != nil {
		fn(m.listSecretsOutput, true)
	}
	return nil
}

func (m *mockSecretsManager) ListSecretsWithContext(ctx aws.Context, input *secretsmanager.ListSecretsInput, opts ...request.Option) (*secretsmanager.ListSecretsOutput, error) {
	if m.listSecretsError != nil {
		return nil, m.listSecretsError
	}
	return m.listSecretsOutput, nil
}

func (m *mockSecretsManager) GetSecretValueWithContext(ctx aws.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
	if m.getSecretValueError != nil {
		return nil, m.getSecretValueError
	}
	return m.getSecretValueOutput, nil
}

func (m *mockSecretsManager) DescribeSecretWithContext(ctx aws.Context, input *secretsmanager.DescribeSecretInput, opts ...request.Option) (*secretsmanager.DescribeSecretOutput, error) {
	if m.describeSecretError != nil {
		return nil, m.describeSecretError
	}
	return m.describeSecretOutput, nil
}

func (m *mockSecretsManager) CreateSecretWithContext(ctx aws.Context, input *secretsmanager.CreateSecretInput, opts ...request.Option) (*secretsmanager.CreateSecretOutput, error) {
	if m.createSecretError != nil {
		return nil, m.createSecretError
	}
	return m.createSecretOutput, nil
}

func (m *mockSecretsManager) UpdateSecretWithContext(ctx aws.Context, input *secretsmanager.UpdateSecretInput, opts ...request.Option) (*secretsmanager.UpdateSecretOutput, error) {
	if m.updateSecretError != nil {
		return nil, m.updateSecretError
	}
	return m.updateSecretOutput, nil
}

func (m *mockSecretsManager) DeleteSecretWithContext(ctx aws.Context, input *secretsmanager.DeleteSecretInput, opts ...request.Option) (*secretsmanager.DeleteSecretOutput, error) {
	if m.deleteSecretError != nil {
		return nil, m.deleteSecretError
	}
	return m.deleteSecretOutput, nil
}

// Helper function to create a test client with a mock secrets manager
func createTestClientWithMock(mockSM secretsmanageriface.SecretsManagerAPI, profile *config.Profile) *AWSClient {
	return &AWSClient{
		client:  mockSM,
		profile: profile,
		logger:  logrus.New(),
		region:  "us-east-1",
		address: "https://secretsmanager.us-east-1.amazonaws.com",
	}
}

func TestAWSClient_Implements(t *testing.T) {
	t.Run("client has all required methods", func(t *testing.T) {
		mockSM := &mockSecretsManager{}
		profile := &config.Profile{}
		client := createTestClientWithMock(mockSM, profile)

		// Verify client implements all required methods
		var _ interface {
			Authenticate() error
			GetAddress() string
			GetStatus(context.Context) (models.ConnectionStatus, error)
			ListSecrets(string) ([]*models.SecretNode, error)
			GetSecret(string) (*models.SecretNode, error)
			CreateSecret(string, map[string]any) error
			UpdateSecret(string, map[string]any) error
			DeleteSecret(string) error
		} = client
	})
}

func TestNewAWSSecretsManagerClient(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name          string
		profile       *config.Profile
		wantError     bool
		errorContains string
	}{
		{
			name: "valid profile with credentials",
			profile: &config.Profile{
				AuthConfig: config.AuthConfig{
					AWSAccessKeyID:     "test-key",
					AWSSecretAccessKey: "test-secret",
					AWSRegion:          "us-west-2",
				},
			},
			wantError: false,
		},
		{
			name: "valid profile with default region",
			profile: &config.Profile{
				AuthConfig: config.AuthConfig{
					AWSAccessKeyID:     "test-key",
					AWSSecretAccessKey: "test-secret",
				},
			},
			wantError: false,
		},
		{
			name: "missing access key",
			profile: &config.Profile{
				AuthConfig: config.AuthConfig{
					AWSSecretAccessKey: "test-secret",
				},
			},
			wantError:     true,
			errorContains: "aws_access_key_id",
		},
		{
			name: "missing secret key",
			profile: &config.Profile{
				AuthConfig: config.AuthConfig{
					AWSAccessKeyID: "test-key",
				},
			},
			wantError:     true,
			errorContains: "aws_secret_access_key",
		},
		{
			name: "with custom endpoint",
			profile: &config.Profile{
				Address: "http://localhost:4566",
				AuthConfig: config.AuthConfig{
					AWSAccessKeyID:     "test-key",
					AWSSecretAccessKey: "test-secret",
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAWSSecretsManagerClient(logger, tt.profile)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, client)
			} else {
				// Note: This will fail if AWS credentials are not configured
				// For production tests, you'd need actual AWS credentials or skip this test
				if err != nil {
					t.Skipf("Skipping test: AWS credentials not configured: %v", err)
					return
				}
				require.NoError(t, err)
				require.NotNil(t, client)
				assert.Equal(t, tt.profile, client.profile)
			}
		})
	}
}

func TestAWSClient_Authenticate(t *testing.T) {
	tests := []struct {
		name          string
		mockSM        *mockSecretsManager
		wantError     bool
		errorContains string
	}{
		{
			name: "successful authentication",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{
					SecretList: []*secretsmanager.SecretListEntry{},
				},
			},
			wantError: false,
		},
		{
			name: "authentication failure",
			mockSM: &mockSecretsManager{
				listSecretsError: errors.New("access denied"),
			},
			wantError:     true,
			errorContains: "failed to authenticate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClientWithMock(tt.mockSM, &config.Profile{})
			err := client.Authenticate()
			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAWSClient_GetAddress(t *testing.T) {
	tests := []struct {
		name   string
		client *AWSClient
		want   string
	}{
		{
			name: "with address set",
			client: &AWSClient{
				address: "aws://123456789:us-east-1",
				profile: &config.Profile{},
			},
			want: "aws://123456789:us-east-1",
		},
		{
			name: "without address, falls back to profile",
			client: &AWSClient{
				address: "",
				profile: &config.Profile{
					Address: "http://localhost:4566",
				},
			},
			want: "http://localhost:4566",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.client.GetAddress()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAWSClient_GetStatus(t *testing.T) {
	ctx := context.Background()
	testTime := time.Now()

	tests := []struct {
		name       string
		mockSM     *mockSecretsManager
		wantStatus models.Status
		wantError  bool
	}{
		{
			name: "connected",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{},
			},
			wantStatus: models.StatusConnected,
			wantError:  false,
		},
		{
			name: "disconnected",
			mockSM: &mockSecretsManager{
				listSecretsError: errors.New("connection failed"),
			},
			wantStatus: models.StatusDisconnected,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClientWithMock(tt.mockSM, &config.Profile{})
			client.region = "us-east-1"
			client.address = "https://secretsmanager.us-east-1.amazonaws.com"

			status, err := client.GetStatus(ctx)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStatus, status.Status)
				if tt.wantStatus == models.StatusConnected {
					assert.Equal(t, "us-east-1", status.ClusterID)
					assert.Equal(t, "AWS Secrets Manager", status.Version)
				}
				assert.WithinDuration(t, testTime, status.LastCheck, time.Second)
			}
		})
	}
}

func TestAWSClient_ListSecrets(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		path       string
		mockSM     *mockSecretsManager
		wantError  bool
		wantCount  int
		wantSecret bool
		wantDir    bool
	}{
		{
			name: "empty list",
			path: "",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{
					SecretList: []*secretsmanager.SecretListEntry{},
				},
			},
			wantError: false,
			wantCount: 0,
		},
		{
			name: "single secret",
			path: "",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{
					SecretList: []*secretsmanager.SecretListEntry{
						{
							Name:        aws.String("test-secret"),
							CreatedDate: &testTime,
						},
					},
				},
			},
			wantError:  false,
			wantCount:  1,
			wantSecret: true,
		},
		{
			name: "nested secrets",
			path: "",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{
					SecretList: []*secretsmanager.SecretListEntry{
						{
							Name:        aws.String("app/db/password"),
							CreatedDate: &testTime,
						},
						{
							Name:        aws.String("app/api/key"),
							CreatedDate: &testTime,
						},
					},
				},
			},
			wantError: false,
			wantCount: 1, // Should return directory "app"
			wantDir:   true,
		},
		{
			name: "filtered by path",
			path: "app",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{
					SecretList: []*secretsmanager.SecretListEntry{
						{
							Name:        aws.String("app/db/password"),
							CreatedDate: &testTime,
						},
						{
							Name:        aws.String("other/secret"),
							CreatedDate: &testTime,
						},
					},
				},
			},
			wantError: false,
			wantCount: 1, // Should return directory "db"
		},
		{
			name: "list error",
			path: "",
			mockSM: &mockSecretsManager{
				listSecretsError: errors.New("list failed"),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClientWithMock(tt.mockSM, &config.Profile{})
			secrets, err := client.ListSecrets(tt.path)
			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, secrets)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, secrets) // secrets can be empty slice but not nil
				assert.Len(t, secrets, tt.wantCount)
				if tt.wantSecret && len(secrets) > 0 {
					assert.True(t, secrets[0].IsSecret)
				}
				if tt.wantDir && len(secrets) > 0 {
					assert.False(t, secrets[0].IsSecret)
				}
			}
		})
	}
}

func TestAWSClient_GetSecret(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		path          string
		mockSM        *mockSecretsManager
		wantError     bool
		wantJSON      bool
		wantPlainText bool
		errorContains string
	}{
		{
			name: "JSON secret",
			path: "test-secret",
			mockSM: &mockSecretsManager{
				getSecretValueOutput: &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(`{"username":"admin","password":"secret123"}`),
				},
				describeSecretOutput: &secretsmanager.DescribeSecretOutput{
					CreatedDate: &testTime,
				},
			},
			wantError: false,
			wantJSON:  true,
		},
		{
			name: "plain text secret",
			path: "test-secret",
			mockSM: &mockSecretsManager{
				getSecretValueOutput: &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String("plain-text-secret"),
				},
				describeSecretOutput: &secretsmanager.DescribeSecretOutput{
					CreatedDate: &testTime,
				},
			},
			wantError:     false,
			wantPlainText: true,
		},
		{
			name: "deleted secret",
			path: "test-secret",
			mockSM: &mockSecretsManager{
				getSecretValueOutput: &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(`{"key":"value"}`),
				},
				describeSecretOutput: &secretsmanager.DescribeSecretOutput{
					CreatedDate: &testTime,
					DeletedDate: &testTime,
				},
			},
			wantError: false,
		},
		{
			name: "get secret error",
			path: "test-secret",
			mockSM: &mockSecretsManager{
				getSecretValueError: errors.New("secret not found"),
			},
			wantError:     true,
			errorContains: "failed to get secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClientWithMock(tt.mockSM, &config.Profile{})
			secret, err := client.GetSecret(tt.path)
			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, secret)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, secret)
				assert.True(t, secret.IsSecret)
				assert.Equal(t, tt.path, secret.Path)
				if tt.wantJSON {
					assert.Contains(t, secret.Data, "username")
					assert.Contains(t, secret.Data, "password")
				}
				if tt.wantPlainText {
					assert.Contains(t, secret.Data, "value")
					assert.Equal(t, "plain-text-secret", secret.Data["value"])
				}
				if tt.mockSM.describeSecretOutput != nil && tt.mockSM.describeSecretOutput.DeletedDate != nil {
					assert.True(t, secret.Metadata.Destroyed)
				}
			}
		})
	}
}

func TestAWSClient_CreateSecret(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		data          map[string]any
		mockSM        *mockSecretsManager
		wantError     bool
		errorContains string
	}{
		{
			name: "successful creation",
			path: "new-secret",
			data: map[string]any{
				"key": "value",
			},
			mockSM: &mockSecretsManager{
				createSecretOutput: &secretsmanager.CreateSecretOutput{},
			},
			wantError: false,
		},
		{
			name: "creation error",
			path: "new-secret",
			data: map[string]any{
				"key": "value",
			},
			mockSM: &mockSecretsManager{
				createSecretError: errors.New("secret already exists"),
			},
			wantError:     true,
			errorContains: "failed to create secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClientWithMock(tt.mockSM, &config.Profile{})
			err := client.CreateSecret(tt.path, tt.data)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAWSClient_UpdateSecret(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		data          map[string]any
		mockSM        *mockSecretsManager
		wantError     bool
		errorContains string
	}{
		{
			name: "successful update",
			path: "existing-secret",
			data: map[string]any{
				"key": "new-value",
			},
			mockSM: &mockSecretsManager{
				updateSecretOutput: &secretsmanager.UpdateSecretOutput{},
			},
			wantError: false,
		},
		{
			name: "update error",
			path: "non-existent-secret",
			data: map[string]any{
				"key": "value",
			},
			mockSM: &mockSecretsManager{
				updateSecretError: errors.New("secret not found"),
			},
			wantError:     true,
			errorContains: "failed to update secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClientWithMock(tt.mockSM, &config.Profile{})
			err := client.UpdateSecret(tt.path, tt.data)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAWSClient_DeleteSecret(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		mockSM        *mockSecretsManager
		wantError     bool
		errorContains string
	}{
		{
			name: "successful deletion",
			path: "secret-to-delete",
			mockSM: &mockSecretsManager{
				deleteSecretOutput: &secretsmanager.DeleteSecretOutput{},
			},
			wantError: false,
		},
		{
			name: "deletion error",
			path: "non-existent-secret",
			mockSM: &mockSecretsManager{
				deleteSecretError: errors.New("secret not found"),
			},
			wantError:     true,
			errorContains: "failed to delete secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClientWithMock(tt.mockSM, &config.Profile{})
			err := client.DeleteSecret(tt.path)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAWSClient_ListSecrets_PathNormalization(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		path      string
		mockSM    *mockSecretsManager
		wantCount int
		wantPath  string
	}{
		{
			name: "path with leading slash",
			path: "/app/db",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{
					SecretList: []*secretsmanager.SecretListEntry{
						{
							Name:        aws.String("app/db/password"),
							CreatedDate: &testTime,
						},
					},
				},
			},
			wantCount: 1,
		},
		{
			name: "path with trailing slash",
			path: "app/db/",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{
					SecretList: []*secretsmanager.SecretListEntry{
						{
							Name:        aws.String("app/db/password"),
							CreatedDate: &testTime,
						},
					},
				},
			},
			wantCount: 1,
		},
		{
			name: "path with both slashes",
			path: "/app/db/",
			mockSM: &mockSecretsManager{
				listSecretsOutput: &secretsmanager.ListSecretsOutput{
					SecretList: []*secretsmanager.SecretListEntry{
						{
							Name:        aws.String("app/db/password"),
							CreatedDate: &testTime,
						},
					},
				},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClientWithMock(tt.mockSM, &config.Profile{})
			secrets, err := client.ListSecrets(tt.path)
			require.NoError(t, err)
			assert.Len(t, secrets, tt.wantCount)
		})
	}
}

func TestAWSClient_GetSecret_PathNormalization(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "path with leading slash",
			path: "/test-secret",
		},
		{
			name: "path with trailing slash",
			path: "test-secret/",
		},
		{
			name: "path with both slashes",
			path: "/test-secret/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSM := &mockSecretsManager{
				getSecretValueOutput: &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(`{"key":"value"}`),
				},
				describeSecretOutput: &secretsmanager.DescribeSecretOutput{
					CreatedDate: &testTime,
				},
			}
			client := createTestClientWithMock(mockSM, &config.Profile{})
			secret, err := client.GetSecret(tt.path)
			require.NoError(t, err)
			require.NotNil(t, secret)
			// Path should be normalized (slashes trimmed)
			assert.Equal(t, "test-secret", secret.Path)
		})
	}
}
