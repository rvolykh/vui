#!/bin/sh -ex

# Check if vault status is already run
if [ -f /status/vault_aws/1 ]; then
    echo "Skipped vault status"
else
    vault status
    touch /status/vault_aws/1
fi

# Enable AWs authentication
if [ -f /status/vault_aws/2 ]; then
    echo "Skipped vault enable aws"
else
    vault auth enable aws
    touch /status/vault_aws/2
fi

# Write AWS configuration
if [ -f /status/vault_aws/3 ]; then
    echo "Skipped vault configure aws"
else
    vault write auth/aws/config/client \
        iam_endpoint="http://localstack:4566" \
        sts_endpoint="http://localstack:4566" \
        access_key="test" \
        secret_key="test"
    touch /status/vault_aws/3
fi

# Write AWS policy
if [ -f /status/vault_aws/4 ]; then
    echo "Skipped vault configure aws policy"
else
    vault policy write rw-aws-policy - <<EOF
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
    touch /status/vault_aws/4
fi

# Write AWS role
if [ -f /status/vault_aws/5 ]; then
    echo "Skipped vault configure aws role"
else
    vault write auth/aws/role/vui-iam-role \
        auth_type=iam \
        bound_iam_principal_arn="arn:aws:iam::000000000000:role/vui-iam-role" \
        policies="rw-aws-policy"
    touch /status/vault_aws/5
fi

# Verify AWS authentication
if [ -f /status/vault_aws/6 ]; then
    echo "Skipped vault login"
else
    eval "$(aws configure export-credentials --profile internal-vui-iam-role --format env)"

    echo "AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID"
    echo "AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY"
    echo "AWS_SESSION_TOKEN: $AWS_SESSION_TOKEN"

    vault login -method=aws \
        role="vui-iam-role" \
        aws_access_key_id="$AWS_ACCESS_KEY_ID" \
        aws_secret_access_key="$AWS_SECRET_ACCESS_KEY" \
        aws_session_token="$AWS_SESSION_TOKEN"
    touch /status/vault_aws/6
fi

echo "Completed"
touch /status/vault_aws/completed
