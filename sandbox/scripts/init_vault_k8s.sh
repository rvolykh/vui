#!/bin/sh -ex

# Check if vault status is already run
if [ -f /status/vault_k8s/1 ]; then
    echo "Skipped vault status"
else
    vault status
    touch /status/vault_k8s/1
fi

# Enable LDAP authentication
if [ -f /status/vault_k8s/2 ]; then
    echo "Skipped vault enable kubernetes"
else
    vault auth enable kubernetes
    touch /status/vault_k8s/2
fi

# Write LDAP configuration
if [ -f /status/vault_k8s/3 ]; then
    echo "Skipped vault configure kubernetes"
else
    sed -i 's#https://127.0.0.1:6443#https://kubernetes.default.svc.cluster.local:6443#g' /root/.kube/config
    KUBE_CA_CERT="$(cat /etc/kubernetes/pki/ca.crt)"
    TOKEN_REVIEW_JWT="$(kubectl get secret vui-sandbox-vault-sa-token -n default -o go-template='{{.data.token}}' | base64 -d)"

    vault write auth/kubernetes/config \
      kubernetes_host="https://kubernetes.default.svc.cluster.local:6443" \
      kubernetes_ca_cert=@/etc/kubernetes/pki/ca.crt \
      token_reviewer_jwt="$TOKEN_REVIEW_JWT" \
      issuer="https://kubernetes.default.svc.cluster.local"
    touch /status/vault_k8s/3
fi

# Create admin policy for Vault
if [ -f /status/vault_k8s/4 ]; then
    echo "Skipped vault create policy"
else
    vault policy write rw-k8s-policy - <<EOF
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
    touch /status/vault_k8s/4
fi

# Attach vault policy to K8s service account
if [ -f /status/vault_k8s/5 ]; then
    echo "Skipped vault attach policy"
else
    vault write auth/kubernetes/role/vui-k8s-role \
        bound_service_account_names=vui-sandbox-app-sa \
        bound_service_account_namespaces=default \
        policies=rw-k8s-policy \
        audience=vault \
        ttl=1h
    touch /status/vault_k8s/5
fi

# Verify K8s authentication
if [ -f /status/vault_k8s/6 ]; then
    echo "Skipped vault login"
else
    USER_JWT="$(kubectl create token vui-sandbox-app-sa --audience vault --duration=1h)"
    vault write auth/kubernetes/login role=vui-k8s-role jwt="$USER_JWT"
    touch /status/vault_k8s/6
fi

echo "Completed"
touch /status/vault_k8s/completed
