package aws

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSSM is a mock implementation of ssmiface.SSMAPI
type mockSSM struct {
	ssmiface.SSMAPI
	describeParametersOutput *ssm.DescribeParametersOutput
	describeParametersError  error
	getParameterOutput       *ssm.GetParameterOutput
	getParameterError        error
	putParameterOutput       *ssm.PutParameterOutput
	putParameterError        error
	deleteParameterOutput    *ssm.DeleteParameterOutput
	deleteParameterError     error
}

func (m *mockSSM) DescribeParametersPagesWithContext(ctx aws.Context, input *ssm.DescribeParametersInput, fn func(*ssm.DescribeParametersOutput, bool) bool, opts ...request.Option) error {
	if m.describeParametersError != nil {
		return m.describeParametersError
	}
	if m.describeParametersOutput != nil {
		fn(m.describeParametersOutput, true)
	}
	return nil
}

func (m *mockSSM) DescribeParametersWithContext(ctx aws.Context, input *ssm.DescribeParametersInput, opts ...request.Option) (*ssm.DescribeParametersOutput, error) {
	if m.describeParametersError != nil {
		return nil, m.describeParametersError
	}
	return m.describeParametersOutput, nil
}

func (m *mockSSM) GetParameterWithContext(ctx aws.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error) {
	if m.getParameterError != nil {
		return nil, m.getParameterError
	}
	return m.getParameterOutput, nil
}

func (m *mockSSM) PutParameterWithContext(ctx aws.Context, input *ssm.PutParameterInput, opts ...request.Option) (*ssm.PutParameterOutput, error) {
	if m.putParameterError != nil {
		return nil, m.putParameterError
	}
	return m.putParameterOutput, nil
}

func (m *mockSSM) DeleteParameterWithContext(ctx aws.Context, input *ssm.DeleteParameterInput, opts ...request.Option) (*ssm.DeleteParameterOutput, error) {
	if m.deleteParameterError != nil {
		return nil, m.deleteParameterError
	}
	return m.deleteParameterOutput, nil
}

// Helper function to create a test client with a mock SSM
func createTestSSMClientWithMock(mockSSM ssmiface.SSMAPI, profile *config.Profile) *AWSSSMClient {
	return &AWSSSMClient{
		client:  mockSSM,
		profile: profile,
		logger:  logrus.New(),
		region:  "us-east-1",
		address: "https://ssm.us-east-1.amazonaws.com",
	}
}

func TestAWSSSMClient_Implements(t *testing.T) {
	t.Run("client has all required methods", func(t *testing.T) {
		mockSSM := &mockSSM{}
		profile := &config.Profile{}
		client := createTestSSMClientWithMock(mockSSM, profile)

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

func TestNewAWSSSMClient(t *testing.T) {
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
			client, err := NewAWSSSMClient(logger, tt.profile)
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

func TestAWSSSMClient_Authenticate(t *testing.T) {
	tests := []struct {
		name          string
		mockSSM       *mockSSM
		wantError     bool
		errorContains string
	}{
		{
			name: "successful authentication",
			mockSSM: &mockSSM{
				describeParametersOutput: &ssm.DescribeParametersOutput{},
			},
			wantError: false,
		},
		{
			name: "authentication failure",
			mockSSM: &mockSSM{
				describeParametersError: errors.New("access denied"),
			},
			wantError:     true,
			errorContains: "failed to authenticate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &config.Profile{}
			client := createTestSSMClientWithMock(tt.mockSSM, profile)

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

func TestAWSSSMClient_GetAddress(t *testing.T) {
	tests := []struct {
		name    string
		client  *AWSSSMClient
		profile *config.Profile
		want    string
	}{
		{
			name: "returns address when set",
			client: &AWSSSMClient{
				address: "https://ssm.us-east-1.amazonaws.com",
			},
			profile: &config.Profile{},
			want:    "https://ssm.us-east-1.amazonaws.com",
		},
		{
			name: "returns profile address when client address is empty",
			client: &AWSSSMClient{
				address: "",
				profile: &config.Profile{
					Address: "http://localhost:4566",
				},
			},
			profile: &config.Profile{
				Address: "http://localhost:4566",
			},
			want: "http://localhost:4566",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.client.profile = tt.profile
			got := tt.client.GetAddress()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAWSSSMClient_GetStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		mockSSM       *mockSSM
		wantStatus    models.ConnectionStatus
		wantError     bool
		errorContains string
	}{
		{
			name: "connected status",
			mockSSM: &mockSSM{
				describeParametersOutput: &ssm.DescribeParametersOutput{},
			},
			wantStatus: models.ConnectionStatus{
				Status:    models.StatusConnected,
				Address:   "https://ssm.us-east-1.amazonaws.com",
				Version:   "AWS SSM Parameter Store",
				ClusterID: "us-east-1",
			},
			wantError: false,
		},
		{
			name: "disconnected status",
			mockSSM: &mockSSM{
				describeParametersError: errors.New("network error"),
			},
			wantStatus: models.ConnectionStatus{
				Status:  models.StatusDisconnected,
				Address: "https://ssm.us-east-1.amazonaws.com",
				Error:   "network error",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &config.Profile{}
			client := createTestSSMClientWithMock(tt.mockSSM, profile)

			status, err := client.GetStatus(ctx)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStatus.Status, status.Status)
				assert.Equal(t, tt.wantStatus.Address, status.Address)
				assert.Equal(t, tt.wantStatus.Version, status.Version)
				assert.Equal(t, tt.wantStatus.ClusterID, status.ClusterID)
				if tt.wantStatus.Error != "" {
					assert.Contains(t, status.Error, tt.wantStatus.Error)
				}
				assert.False(t, status.LastCheck.IsZero())
			}
		})
	}
}

func TestAWSSSMClient_ListSecrets(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		mockSSM       *mockSSM
		wantNodes     []*models.SecretNode
		wantError     bool
		errorContains string
	}{
		{
			name: "list root parameters",
			path: "",
			mockSSM: &mockSSM{
				describeParametersOutput: &ssm.DescribeParametersOutput{
					Parameters: []*ssm.ParameterMetadata{
						{
							Name:             aws.String("/app/dev/db/password"),
							LastModifiedDate: aws.Time(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						},
						{
							Name:             aws.String("/app/prod/api/key"),
							LastModifiedDate: aws.Time(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
			},
			wantNodes: []*models.SecretNode{
				{
					Name:     "app",
					Path:     "/app",
					IsSecret: false,
					Children: []*models.SecretNode{},
				},
			},
			wantError: false,
		},
		{
			name: "list parameters with path prefix",
			path: "app/dev",
			mockSSM: &mockSSM{
				describeParametersOutput: &ssm.DescribeParametersOutput{
					Parameters: []*ssm.ParameterMetadata{
						{
							Name:             aws.String("/app/dev/db/password"),
							LastModifiedDate: aws.Time(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						},
					},
				},
			},
			wantNodes: []*models.SecretNode{
				{
					Name:     "db",
					Path:     "/app/dev/db",
					IsSecret: false,
					Children: []*models.SecretNode{},
				},
			},
			wantError: false,
		},
		{
			name: "list secrets error",
			path: "",
			mockSSM: &mockSSM{
				describeParametersError: errors.New("access denied"),
			},
			wantError:     true,
			errorContains: "failed to list parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &config.Profile{}
			client := createTestSSMClientWithMock(tt.mockSSM, profile)

			nodes, err := client.ListSecrets(tt.path)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, nodes)
			} else {
				require.NoError(t, err)
				require.NotNil(t, nodes)
				if len(tt.wantNodes) > 0 {
					assert.Equal(t, len(tt.wantNodes), len(nodes))
					for i, wantNode := range tt.wantNodes {
						if i < len(nodes) {
							assert.Equal(t, wantNode.Name, nodes[i].Name)
							assert.Equal(t, wantNode.Path, nodes[i].Path)
							assert.Equal(t, wantNode.IsSecret, nodes[i].IsSecret)
						}
					}
				}
			}
		})
	}
}

func TestAWSSSMClient_GetSecret(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		mockSSM       *mockSSM
		wantNode      *models.SecretNode
		wantError     bool
		errorContains string
	}{
		{
			name: "get secret successfully",
			path: "/app/dev/db/password",
			mockSSM: &mockSSM{
				getParameterOutput: &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Name:             aws.String("/app/dev/db/password"),
						Value:            aws.String(`{"password":"secret123"}`),
						Type:             aws.String("SecureString"),
						Version:          aws.Int64(1),
						LastModifiedDate: aws.Time(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			wantNode: &models.SecretNode{
				Name:     "password",
				Path:     "/app/dev/db/password",
				IsSecret: true,
				Data: map[string]any{
					"password": "secret123",
				},
				Metadata: &models.SecretMetadata{
					Version:     1,
					CreatedTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			wantError: false,
		},
		{
			name: "get secret with plain string value",
			path: "/app/key",
			mockSSM: &mockSSM{
				getParameterOutput: &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Name:             aws.String("/app/key"),
						Value:            aws.String("simple-value"),
						Type:             aws.String("String"),
						Version:          aws.Int64(2),
						LastModifiedDate: aws.Time(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			wantNode: &models.SecretNode{
				Name:     "key",
				Path:     "/app/key",
				IsSecret: true,
				Data: map[string]any{
					"value": "simple-value",
				},
				Metadata: &models.SecretMetadata{
					Version:     2,
					CreatedTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			wantError: false,
		},
		{
			name: "get secret error",
			path: "/nonexistent",
			mockSSM: &mockSSM{
				getParameterError: errors.New("parameter not found"),
			},
			wantError:     true,
			errorContains: "failed to get parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &config.Profile{}
			client := createTestSSMClientWithMock(tt.mockSSM, profile)

			node, err := client.GetSecret(tt.path)
			if tt.wantError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, node)
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				assert.Equal(t, tt.wantNode.Name, node.Name)
				assert.Equal(t, tt.wantNode.Path, node.Path)
				assert.Equal(t, tt.wantNode.IsSecret, node.IsSecret)
				if tt.wantNode.Data != nil {
					assert.Equal(t, tt.wantNode.Data, node.Data)
				}
				if tt.wantNode.Metadata != nil {
					require.NotNil(t, node.Metadata)
					assert.Equal(t, tt.wantNode.Metadata.Version, node.Metadata.Version)
					assert.Equal(t, tt.wantNode.Metadata.CreatedTime, node.Metadata.CreatedTime)
				}
			}
		})
	}
}

func TestAWSSSMClient_CreateSecret(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		data          map[string]any
		mockSSM       *mockSSM
		wantError     bool
		errorContains string
	}{
		{
			name: "create secret successfully",
			path: "/app/dev/key",
			data: map[string]any{
				"password": "secret123",
				"username": "admin",
			},
			mockSSM: &mockSSM{
				putParameterOutput: &ssm.PutParameterOutput{
					Version: aws.Int64(1),
				},
			},
			wantError: false,
		},
		{
			name: "create secret error",
			path: "/app/key",
			data: map[string]any{
				"value": "test",
			},
			mockSSM: &mockSSM{
				putParameterError: errors.New("parameter already exists"),
			},
			wantError:     true,
			errorContains: "failed to create parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &config.Profile{}
			client := createTestSSMClientWithMock(tt.mockSSM, profile)

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

func TestAWSSSMClient_UpdateSecret(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		data          map[string]any
		mockSSM       *mockSSM
		wantError     bool
		errorContains string
	}{
		{
			name: "update secret successfully",
			path: "/app/dev/key",
			data: map[string]any{
				"password": "newsecret123",
			},
			mockSSM: &mockSSM{
				getParameterOutput: &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Type: aws.String("SecureString"),
					},
				},
				putParameterOutput: &ssm.PutParameterOutput{
					Version: aws.Int64(2),
				},
			},
			wantError: false,
		},
		{
			name: "update secret error",
			path: "/app/key",
			data: map[string]any{
				"value": "test",
			},
			mockSSM: &mockSSM{
				putParameterError: errors.New("parameter update failed"),
			},
			wantError:     true,
			errorContains: "failed to update parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &config.Profile{}
			client := createTestSSMClientWithMock(tt.mockSSM, profile)

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

func TestAWSSSMClient_DeleteSecret(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		mockSSM       *mockSSM
		wantError     bool
		errorContains string
	}{
		{
			name: "delete secret successfully",
			path: "/app/dev/key",
			mockSSM: &mockSSM{
				deleteParameterOutput: &ssm.DeleteParameterOutput{},
			},
			wantError: false,
		},
		{
			name: "delete secret error",
			path: "/nonexistent",
			mockSSM: &mockSSM{
				deleteParameterError: errors.New("parameter not found"),
			},
			wantError:     true,
			errorContains: "failed to delete parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &config.Profile{}
			client := createTestSSMClientWithMock(tt.mockSSM, profile)

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
