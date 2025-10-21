#!/bin/sh -ex

# Check if vault status is already run
if [ -f /status/vault_jwt/1 ]; then
    echo "Skipped vault status"
else
    vault status
    touch /status/vault_jwt/1
fi

# Enable LDAP authentication
if [ -f /status/vault_jwt/2 ]; then
    echo "Skipped vault enable jwt"
else
    vault auth enable jwt
    touch /status/vault_jwt/2
fi

# Create admin policy for Vault
if [ -f /status/vault_jwt/3 ]; then
    echo "Skipped vault create policy"
else
    vault policy write rw-jwt-policy - <<EOF
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
    touch /status/vault_jwt/3
fi

# Attach vault policy to JWT role
if [ -f /status/vault_jwt/4 ]; then
    echo "Skipped vault attach policy"
else
    vault write auth/jwt/role/vui-jwt-role - <<EOF
    {
      "role_type": "jwt",
      "policies": ["rw-jwt-policy"],
      "user_claim": "email",
      "bound_claims": {
        "email": "vui-sandbox-user@localhost"
      }
    }
EOF
    touch /status/vault_jwt/4
fi

# Write JWT configuration
if [ -f /status/vault_jwt/5 ]; then
    echo "Skipped vault configure jwt"
else
    vault write auth/jwt/config \
      oidc_discovery_url="http://oidc:8080/realms/vui-sandbox-realm" \
      oidc_client_id="" \
      oidc_client_secret="" \
      default_role="vui-jwt-role"
    touch /status/vault_jwt/5
fi

echo "Completed"
touch /status/vault_jwt/completed
