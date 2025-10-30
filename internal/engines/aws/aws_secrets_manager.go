package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/models"
	"github.com/rvolykh/vui/internal/utils"
	"github.com/sirupsen/logrus"
)

type AWSClient struct {
	client  secretsmanageriface.SecretsManagerAPI
	profile *config.Profile
	logger  *logrus.Logger
	region  string
	address string
}

func NewAWSSecretsManagerClient(logger *logrus.Logger, profile *config.Profile) (*AWSClient, error) {
	region := utils.Coalesce(profile.AuthConfig.AWSRegion, "us-east-1")

	awsConfig := aws.NewConfig().WithRegion(region)

	if profile.Address != "" {
		awsConfig.WithEndpoint(profile.Address)
	}

	if profile.AuthConfig.AWSAccessKeyID == "" || profile.AuthConfig.AWSSecretAccessKey == "" {
		return nil, fmt.Errorf("aws_access_key_id and aws_secret_access_key are required for AWS Secrets Manager authentication")
	}

	awsConfig.WithCredentials(credentials.NewStaticCredentials(
		profile.AuthConfig.AWSAccessKeyID,
		profile.AuthConfig.AWSSecretAccessKey,
		profile.AuthConfig.AWSSessionToken,
	))

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session for static credentials: %w", err)
	}

	awsRole := profile.AuthConfig.AWSRole
	if awsRole != "" {
		awsConfig.WithCredentials(stscreds.NewCredentials(sess, awsRole))

		sess, err = session.NewSession(awsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS session for assumed role: %w", err)
		}
	}

	address := profile.Address
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := sts.New(sess).GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
	if err == nil && result != nil && result.Account != nil {
		address = fmt.Sprintf("aws://%s:%s", *result.Account, region)
	} else if profile.Address == "" {
		address = fmt.Sprintf("secretsmanager.%s.amazonaws.com", region)
	}

	return &AWSClient{
		client:  secretsmanager.New(sess),
		profile: profile,
		logger:  logger,
		region:  region,
		address: address,
	}, nil
}

func (c *AWSClient) Authenticate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// For AWS Secrets Manager, we can verify by making a simple API call
	_, err := c.client.ListSecretsWithContext(ctx, &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int64(1),
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with AWS Secrets Manager: %w", err)
	}

	c.logger.Debug("AWS Secrets Manager authentication verified successfully")
	return nil
}

func (c *AWSClient) GetAddress() string {
	if c.address == "" {
		return c.profile.Address
	}
	return c.address
}

func (c *AWSClient) GetStatus(ctx context.Context) (models.ConnectionStatus, error) {
	_, err := c.client.ListSecretsWithContext(ctx, &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int64(1),
	})
	if err != nil {
		return models.ConnectionStatus{
			Status:    models.StatusDisconnected,
			Address:   c.GetAddress(),
			LastCheck: time.Now(),
			Error:     err.Error(),
		}, nil
	}

	// Get AWS account ID or region info for cluster_id
	// For AWS, we can use the region as cluster identifier
	return models.ConnectionStatus{
		Status:    models.StatusConnected,
		Address:   c.GetAddress(),
		Version:   "AWS Secrets Manager",
		ClusterID: c.region,
		LastCheck: time.Now(),
	}, nil
}

func (c *AWSClient) ListSecrets(path string) ([]*models.SecretNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// AWS Secrets Manager doesn't have a hierarchical path structure like Vault
	// We'll use prefix matching to simulate directories
	input := &secretsmanager.ListSecretsInput{}

	normalizedPath := strings.Trim(path, "/")

	var allSecrets []*secretsmanager.SecretListEntry
	err := c.client.ListSecretsPagesWithContext(ctx, input, func(page *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		if page.SecretList != nil {
			allSecrets = append(allSecrets, page.SecretList...)
		}
		return !lastPage
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	// Filter secrets by path prefix if a path is specified
	if normalizedPath != "" {
		filteredSecrets := []*secretsmanager.SecretListEntry{}
		prefix := normalizedPath + "/"
		for _, secret := range allSecrets {
			if secret.Name != nil && strings.HasPrefix(*secret.Name, prefix) {
				filteredSecrets = append(filteredSecrets, secret)
			}
		}
		allSecrets = filteredSecrets
	}

	// Build a tree structure from secret names
	// Secrets can have "/" in their names to simulate paths
	// We'll return immediate children (both secrets and directories) at the current path level
	nodeMap := make(map[string]*models.SecretNode)

	for _, secret := range allSecrets {
		if secret.Name == nil {
			continue
		}

		secretName := *secret.Name
		// Remove the prefix if we're filtering by path
		var relativePath string
		if normalizedPath != "" {
			if !strings.HasPrefix(secretName, normalizedPath+"/") {
				continue
			}
			relativePath = strings.TrimPrefix(secretName, normalizedPath+"/")
		} else {
			relativePath = secretName
		}

		// Split into parts to find immediate children
		parts := strings.Split(relativePath, "/")
		if len(parts) == 0 {
			continue
		}

		firstPart := parts[0]
		fullPath := firstPart
		if normalizedPath != "" {
			fullPath = normalizedPath + "/" + firstPart
		}

		// Check if this is a direct child or a nested path
		if len(parts) == 1 {
			// This is a direct secret at the current path level
			node := &models.SecretNode{
				Name:     firstPart,
				Path:     secretName, // Use full secret name as path
				IsSecret: true,
			}

			if secret.CreatedDate != nil {
				node.Metadata = &models.SecretMetadata{
					CreatedTime: *secret.CreatedDate,
					Version:     1,
				}
			}

			nodeMap[firstPart] = node
		} else {
			// This is a nested path, create a directory node
			if _, exists := nodeMap[firstPart]; !exists {
				dirNode := &models.SecretNode{
					Name:     firstPart,
					Path:     fullPath,
					IsSecret: false,
					Children: []*models.SecretNode{},
				}
				nodeMap[firstPart] = dirNode
			}
		}
	}

	// Convert map to slice
	result := []*models.SecretNode{}
	for _, node := range nodeMap {
		result = append(result, node)
	}

	return result, nil
}

func (c *AWSClient) GetSecret(path string) (*models.SecretNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// AWS Secrets Manager uses secret name or ARN as identifier
	secretName := strings.Trim(path, "/")

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := c.client.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret '%s': %w", path, err)
	}

	node := &models.SecretNode{
		Name:     filepath.Base(secretName),
		Path:     secretName,
		IsSecret: true,
		Metadata: &models.SecretMetadata{},
	}

	// Parse secret value (can be JSON string or plain string)
	var secretData map[string]any
	secretString := ""
	if result.SecretString != nil {
		secretString = *result.SecretString
	}

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		// If not JSON, treat as plain string
		secretData = map[string]any{
			"value": secretString,
		}
	}

	node.Data = secretData

	// Get metadata
	describeInput := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretName),
	}
	describeResult, err := c.client.DescribeSecretWithContext(ctx, describeInput)
	if err == nil && describeResult != nil {
		if describeResult.CreatedDate != nil {
			node.Metadata.CreatedTime = *describeResult.CreatedDate
		}
		// AWS Secrets Manager has versioning but not a simple integer version
		// We'll use 1 as default version
		node.Metadata.Version = 1
		if describeResult.DeletedDate != nil {
			node.Metadata.DeletionTime = *describeResult.DeletedDate
			node.Metadata.Destroyed = true
		}
	}

	return node, nil
}

func (c *AWSClient) CreateSecret(path string, data map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretName := strings.Trim(path, "/")

	var secretString string
	// Check if data has empty key marker (omitted key case)
	if value, hasEmptyKey := data[""]; hasEmptyKey && len(data) == 1 {
		// Store value as plain string (no JSON conversion)
		secretString = fmt.Sprintf("%v", value)
	} else {
		// Convert data map to JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal secret data: %w", err)
		}
		secretString = string(jsonData)
	}

	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretString),
	}

	_, err := c.client.CreateSecretWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create secret '%s': %w", path, err)
	}

	c.logger.Infof("Created secret: %s", secretName)
	return nil
}

func (c *AWSClient) UpdateSecret(path string, data map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretName := strings.Trim(path, "/")

	var secretString string
	// Check if data has empty key marker (omitted key case)
	if value, hasEmptyKey := data[""]; hasEmptyKey && len(data) == 1 {
		// Store value as plain string (no JSON conversion)
		secretString = fmt.Sprintf("%v", value)
	} else {
		// Convert data map to JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal secret data: %w", err)
		}
		secretString = string(jsonData)
	}

	input := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(secretString),
	}

	_, err := c.client.UpdateSecretWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update secret '%s': %w", path, err)
	}

	c.logger.Infof("Updated secret: %s", secretName)
	return nil
}

func (c *AWSClient) DeleteSecret(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretName := strings.Trim(path, "/")

	input := &secretsmanager.DeleteSecretInput{
		SecretId: aws.String(secretName),
	}

	_, err := c.client.DeleteSecretWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete secret '%s': %w", path, err)
	}

	c.logger.Infof("Deleted secret: %s", secretName)
	return nil
}
