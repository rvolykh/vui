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
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/rvolykh/vui/internal/config"
	"github.com/rvolykh/vui/internal/models"
	"github.com/rvolykh/vui/internal/utils"
	"github.com/sirupsen/logrus"
)

type AWSSSMClient struct {
	client  ssmiface.SSMAPI
	profile *config.Profile
	logger  *logrus.Logger
	region  string
	address string
}

func NewAWSSSMClient(logger *logrus.Logger, profile *config.Profile) (*AWSSSMClient, error) {
	region := utils.Coalesce(profile.AuthConfig.AWSRegion, "us-east-1")

	awsConfig := aws.NewConfig().WithRegion(region)

	if profile.Address != "" {
		awsConfig.WithEndpoint(profile.Address)
	}

	if profile.AuthConfig.AWSAccessKeyID == "" || profile.AuthConfig.AWSSecretAccessKey == "" {
		return nil, fmt.Errorf("aws_access_key_id and aws_secret_access_key are required for AWS SSM Parameter Store authentication")
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
		address = fmt.Sprintf("ssm.%s.amazonaws.com", region)
	}

	return &AWSSSMClient{
		client:  ssm.New(sess),
		profile: profile,
		logger:  logger,
		region:  region,
		address: address,
	}, nil
}

func (c *AWSSSMClient) Authenticate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// For AWS SSM, we can verify by making a simple API call
	_, err := c.client.DescribeParametersWithContext(ctx, &ssm.DescribeParametersInput{
		MaxResults: aws.Int64(1),
	})
	if err != nil {
		return fmt.Errorf("failed to authenticate with AWS SSM Parameter Store: %w", err)
	}

	c.logger.Debug("AWS SSM Parameter Store authentication verified successfully")
	return nil
}

func (c *AWSSSMClient) GetAddress() string {
	if c.address == "" {
		return c.profile.Address
	}
	return c.address
}

func (c *AWSSSMClient) GetStatus(ctx context.Context) (models.ConnectionStatus, error) {
	_, err := c.client.DescribeParametersWithContext(ctx, &ssm.DescribeParametersInput{
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

	return models.ConnectionStatus{
		Status:    models.StatusConnected,
		Address:   c.GetAddress(),
		Version:   "AWS SSM Parameter Store",
		ClusterID: c.region,
		LastCheck: time.Now(),
	}, nil
}

func (c *AWSSSMClient) ListSecrets(path string) ([]*models.SecretNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	normalizedPath := strings.Trim(path, "/")

	// SSM Parameter Store uses hierarchical paths
	// If path is empty, we start from root. Otherwise, we use the path as a prefix
	parameterFilters := []*ssm.ParameterStringFilter{}

	if normalizedPath != "" {
		// Add a prefix filter - SSM paths typically start with /
		prefix := "/" + normalizedPath
		parameterFilters = append(parameterFilters, &ssm.ParameterStringFilter{
			Key:    aws.String("Name"),
			Option: aws.String("BeginsWith"),
			Values: []*string{aws.String(prefix)},
		})
	}

	input := &ssm.DescribeParametersInput{
		MaxResults:       aws.Int64(10),
		ParameterFilters: parameterFilters,
	}

	var allParameters []*ssm.ParameterMetadata
	err := c.client.DescribeParametersPagesWithContext(ctx, input, func(page *ssm.DescribeParametersOutput, lastPage bool) bool {
		if page.Parameters != nil {
			allParameters = append(allParameters, page.Parameters...)
		}
		return !lastPage
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list parameters: %w", err)
	}

	// Build a tree structure from parameter names
	nodeMap := make(map[string]*models.SecretNode)

	for _, param := range allParameters {
		if param.Name == nil {
			continue
		}

		paramName := *param.Name
		// Remove leading slash if present
		paramName = strings.TrimPrefix(paramName, "/")

		// Remove the prefix if we're filtering by path
		var relativePath string
		if normalizedPath != "" {
			if !strings.HasPrefix(paramName, normalizedPath+"/") && paramName != normalizedPath {
				continue
			}
			if paramName == normalizedPath {
				// This is the exact path requested, skip it as it's not a child
				continue
			}
			relativePath = strings.TrimPrefix(paramName, normalizedPath+"/")
		} else {
			relativePath = paramName
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
		// Ensure full path starts with /
		if !strings.HasPrefix(fullPath, "/") {
			fullPath = "/" + fullPath
		}

		// Check if this is a direct child or a nested path
		if len(parts) == 1 {
			// This is a direct parameter at the current path level
			node := &models.SecretNode{
				Name:     firstPart,
				Path:     "/" + paramName, // Use full parameter name with leading slash
				IsSecret: true,
			}

			if param.LastModifiedDate != nil {
				node.Metadata = &models.SecretMetadata{
					CreatedTime: *param.LastModifiedDate,
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

func (c *AWSSSMClient) GetSecret(path string) (*models.SecretNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// AWS SSM uses parameter name as identifier
	// Ensure path starts with /
	paramName := strings.Trim(path, "/")
	if !strings.HasPrefix(paramName, "/") {
		paramName = "/" + paramName
	}

	input := &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true), // Always decrypt SecureString parameters
	}

	result, err := c.client.GetParameterWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get parameter '%s': %w", path, err)
	}

	if result.Parameter == nil {
		return nil, fmt.Errorf("parameter '%s' not found", path)
	}

	param := result.Parameter
	node := &models.SecretNode{
		Name:     filepath.Base(paramName),
		Path:     paramName,
		IsSecret: true,
		Metadata: &models.SecretMetadata{},
	}

	// Parse parameter value
	var secretData map[string]any
	paramValue := ""
	if param.Value != nil {
		paramValue = *param.Value
	}

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(paramValue), &secretData); err != nil {
		// If not JSON, treat as plain string
		secretData = map[string]any{
			"value": paramValue,
		}
	}

	node.Data = secretData

	// Get additional metadata
	if param.Version != nil {
		node.Metadata.Version = int(*param.Version)
	}
	if param.LastModifiedDate != nil {
		node.Metadata.CreatedTime = *param.LastModifiedDate
	}

	return node, nil
}

func (c *AWSSSMClient) CreateSecret(path string, data map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure path starts with /
	paramName := strings.Trim(path, "/")
	if !strings.HasPrefix(paramName, "/") {
		paramName = "/" + paramName
	}

	// Convert data map to JSON string
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal secret data: %w", err)
	}

	// Determine parameter type - if data contains sensitive info, use SecureString
	// For simplicity, we'll use SecureString for all secrets
	paramType := "SecureString"

	// Check if the data suggests it's a plain string (single "value" key)
	if len(data) == 1 {
		if _, ok := data["value"]; ok {
			// If it's a simple value, use String type
			paramType = "String"
			jsonData = []byte(fmt.Sprintf("%v", data["value"]))
		}
	}

	input := &ssm.PutParameterInput{
		Name:      aws.String(paramName),
		Value:     aws.String(string(jsonData)),
		Type:      aws.String(paramType),
		Overwrite: aws.Bool(false), // Don't overwrite existing parameters
	}

	_, err = c.client.PutParameterWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create parameter '%s': %w", path, err)
	}

	c.logger.Infof("Created parameter: %s", paramName)
	return nil
}

func (c *AWSSSMClient) UpdateSecret(path string, data map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure path starts with /
	paramName := strings.Trim(path, "/")
	if !strings.HasPrefix(paramName, "/") {
		paramName = "/" + paramName
	}

	// Convert data map to JSON string
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal secret data: %w", err)
	}

	// Get existing parameter to determine type
	getInput := &ssm.GetParameterInput{
		Name: aws.String(paramName),
	}
	existingParam, err := c.client.GetParameterWithContext(ctx, getInput)

	paramType := "SecureString"
	if err == nil && existingParam != nil && existingParam.Parameter != nil && existingParam.Parameter.Type != nil {
		paramType = *existingParam.Parameter.Type
	} else {
		// If parameter doesn't exist or we can't determine type, check data structure
		if len(data) == 1 {
			if _, ok := data["value"]; ok {
				paramType = "String"
				jsonData = []byte(fmt.Sprintf("%v", data["value"]))
			}
		}
	}

	input := &ssm.PutParameterInput{
		Name:      aws.String(paramName),
		Value:     aws.String(string(jsonData)),
		Type:      aws.String(paramType),
		Overwrite: aws.Bool(true), // Overwrite for updates
	}

	_, err = c.client.PutParameterWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update parameter '%s': %w", path, err)
	}

	c.logger.Infof("Updated parameter: %s", paramName)
	return nil
}

func (c *AWSSSMClient) DeleteSecret(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure path starts with /
	paramName := strings.Trim(path, "/")
	if !strings.HasPrefix(paramName, "/") {
		paramName = "/" + paramName
	}

	input := &ssm.DeleteParameterInput{
		Name: aws.String(paramName),
	}

	_, err := c.client.DeleteParameterWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete parameter '%s': %w", path, err)
	}

	c.logger.Infof("Deleted parameter: %s", paramName)
	return nil
}
