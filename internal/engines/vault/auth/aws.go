package auth

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/go-secure-stdlib/awsutil"
	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/adapters"
	"github.com/rvolykh/vui/internal/config"
)

// authenticateWithAWS authenticates using AWS
func (am *AuthManager) authenticateWithAWS(client *api.Client, profile *config.Profile) error {
	accessKeyID := profile.AuthConfig.AWSAccessKeyID
	if accessKeyID == "" {
		return fmt.Errorf("aws_access_key_id is required for AWS authentication")
	}

	secretAccessKey := profile.AuthConfig.AWSSecretAccessKey
	if secretAccessKey == "" {
		return fmt.Errorf("aws_secret_access_key is required for AWS authentication")
	}

	sessionToken := profile.AuthConfig.AWSSessionToken

	role := profile.AuthConfig.AWSRole
	if role == "" {
		return fmt.Errorf("aws_role is required for AWS authentication")
	}

	region := profile.AuthConfig.AWSRegion
	if region == "" {
		region = "us-east-1"
	}

	creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, sessionToken)

	data, err := awsutil.GenerateLoginData(creds, "", region, adapters.NewHclogAdapter(am.logger))
	if err != nil {
		return fmt.Errorf("unable to generate login data for AWS auth endpoint: %w", err)
	}

	secret, err := client.Logical().Write("auth/aws/login", data)
	if err != nil {
		return fmt.Errorf("failed to authenticate with AWS: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no authentication data returned from AWS")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}
