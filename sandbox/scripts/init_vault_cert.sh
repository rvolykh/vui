#!/bin/sh -ex

# Check if vault status is already run
if [ -f /status/vault_cert/1 ]; then
    echo "Skipped vault status"
else
    vault status
    touch /status/vault_cert/1
fi

# Enable CERT authentication
if [ -f /status/vault_cert/2 ]; then
    echo "Skipped vault enable cert"
else
    vault auth enable cert
    touch /status/vault_cert/2
fi

# Create admin policy for Vault
if [ -f /status/vault_cert/3 ]; then
    echo "Skipped vault create policy"
else
    vault policy write rw-cert-policy - <<EOF
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
    touch /status/vault_cert/3
fi

# Create admin policy for Vault
if [ -f /status/vault_cert/4 ]; then
    echo "Skipped vault create policy"
else
    echo "Generating CA private key and self-signed certificate..."
    openssl genrsa -out /certs/ca.key 2048
    openssl req -new -x509 -key /certs/ca.key -out /certs/ca.crt -days 365 -subj "/CN=VUI CA"

    echo "Generating client private key and CSR..."
    openssl genrsa -out /certs/client.key 2048
    openssl req -new -key /certs/client.key -out /certs/client.csr -subj "/CN=vui-sandbox-user"

    echo "extendedKeyUsage=clientAuth" > /certs/extfile.cnf

    echo "Signing the client certificate with the CA..."
    openssl x509 -req -in /certs/client.csr -CA /certs/ca.crt -CAkey /certs/ca.key -CAcreateserial \
        -out /certs/client.crt -days 365 -extfile /certs/extfile.cnf

    touch /status/vault_cert/4
fi

# Write CERT configuration
if [ -f /status/vault_cert/5 ]; then
    echo "Skipped vault configure oidc"
else
    vault write auth/cert/certs/vui-sandbox-user \
        display_name=vui \
        policies=rw-cert-policy \
        certificate=@/certs/client.crt \
        ttl=3600
    touch /status/vault_cert/5
fi

# Verify CERT authentication
if [ -f /status/vault_cert/6 ]; then
    echo "Skipped vault login"
else
    vault login -method=cert -client-cert=/certs/client.crt -client-key=/certs/client.key
    touch /status/vault_cert/6
fi

echo "Completed"
touch /status/vault_cert/completed
