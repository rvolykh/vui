#!/bin/sh -ex

# Check if vault status is already run
if [ -f /status/vault_oidc/1 ]; then
    echo "Skipped vault status"
else
    vault status
    touch /status/vault_oidc/1
fi

# Enable OIDC authentication
if [ -f /status/vault_oidc/2 ]; then
    echo "Skipped vault enable oidc"
else
    vault auth enable oidc
    touch /status/vault_oidc/2
fi

# Create admin policy for Vault
if [ -f /status/vault_oidc/3 ]; then
    echo "Skipped vault create policy"
else
    vault policy write rw-oidc-policy - <<EOF
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
    touch /status/vault_oidc/3
fi

# Attach vault policy to OIDC role
if [ -f /status/vault_oidc/4 ]; then
    echo "Skipped vault attach policy"
else
    vault write auth/oidc/role/vui-oidc-role \
        bound_audiences="vui-sandbox-oidc-client-id" \
        allowed_redirect_uris="http://localhost:8250/oidc/callback" \
        policies=rw-oidc-policy \
        user_claim=sub \
        role_type=oidc \
        ttl=1h
    touch /status/vault_oidc/4
fi

# Write OIDC configuration
if [ -f /status/vault_oidc/5 ]; then
    echo "Skipped vault configure oidc"
else
    vault write auth/oidc/config \
      oidc_discovery_url="http://oidc:8080/realms/vui-sandbox-realm" \
      oidc_client_id="vui-sandbox-oidc-client-id" \
      oidc_client_secret="vui-sandbox-oidc-client-secret" \
      default_role="vui-oidc-role"
    touch /status/vault_oidc/5
fi

echo "Completed"
touch /status/vault_oidc/completed
