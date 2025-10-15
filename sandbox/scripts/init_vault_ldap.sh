#!/bin/sh -ex

# Check if vault status is already run
if [ -f /status/vault_ldap/1 ]; then
    echo "Skipped vault status"
else
    vault status
    touch /status/vault_ldap/1
fi

# Enable LDAP authentication
if [ -f /status/vault_ldap/2 ]; then
    echo "Skipped vault enable ldap"
else
    vault auth enable ldap
    touch /status/vault_ldap/2
fi

# Write LDAP configuration
if [ -f /status/vault_ldap/3 ]; then
    echo "Skipped vault configure ldap"
else
    vault write auth/ldap/config \
      url="ldap://ldap:389" \
      binddn="cn=admin,dc=sandbox,dc=vui" \
      bindpass="vui-sandbox-admin-password" \
      userdn="ou=user,dc=sandbox,dc=vui" \
      groupdn="ou=groups,dc=sandbox,dc=vui" \
      userattr="uid" \
      groupattr="cn" \
      insecure_tls=true
    touch /status/vault_ldap/3
fi

# Create admin policy for Vault
if [ -f /status/vault_ldap/4 ]; then
    echo "Skipped vault create policy"
else
    vault policy write rw-policy - <<EOF
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
    touch /status/vault_ldap/4
fi

# Attach vault policy to LDAP group
if [ -f /status/vault_ldap/5 ]; then
    echo "Skipped vault attach policy"
else
    vault write auth/ldap/groups/testgroup policies=rw-policy
    touch /status/vault_ldap/5
fi

# Verify LDAP authentication
if [ -f /status/vault_ldap/6 ]; then
    echo "Skipped vault login"
else
    vault login -method=ldap username=testuser password=testpassword
    touch /status/vault_ldap/6
fi

echo "Completed"
touch /status/vault_ldap/completed
