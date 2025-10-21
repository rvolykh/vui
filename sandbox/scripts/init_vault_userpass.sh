#!/bin/sh -ex

# Check if vault status is already run
if [ -f /status/vault_userpass/1 ]; then
    echo "Skipped vault status"
else
    vault status
    touch /status/vault_userpass/1
fi

# Enable UserPass authentication
if [ -f /status/vault_userpass/2 ]; then
    echo "Skipped vault enable userpass"
else
    vault auth enable userpass
    touch /status/vault_userpass/2
fi

# Create admin policy for Vault
if [ -f /status/vault_userpass/3 ]; then
    echo "Skipped vault create policy"
else
    vault policy write rw-userpass-policy - <<EOF
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
    touch /status/vault_userpass/3
fi

# Create UserPass user
if [ -f /status/vault_userpass/4 ]; then
    echo "Skipped vault create user"
else
    vault write auth/userpass/users/vui \
        password=vui \
        policies=rw-userpass-policy
    touch /status/vault_userpass/4
fi

# Verify UserPass authentication
if [ -f /status/vault_userpass/5 ]; then
    echo "Skipped vault login"
else
    vault login -method=userpass username=vui password=vui
    touch /status/vault_userpass/5
fi

echo "Completed"
touch /status/vault_userpass/completed
