# VUI Sandbox Environment

## Purpose

The VUI Sandbox is a **local development and testing environment** that provides a complete ecosystem for testing VUI against multiple HashiCorp Vault authentication methods. It eliminates the need for external dependencies and cloud resources by simulating a production-like Vault setup locally using Docker containers and Kubernetes (kind).

### Key Benefits

- **Comprehensive Testing**: Test all supported Vault authentication methods in one environment
- **No Cloud Dependencies**: Everything runs locally - no AWS, Azure, or other cloud accounts needed
- **Reproducible**: Consistent environment setup across different machines
- **Safe Experimentation**: Isolated environment that can be torn down and rebuilt quickly
- **Development Efficiency**: Pre-configured profiles and credentials reduce setup time

---

## Architecture Overview

The sandbox environment consists of several interconnected services:

```
┌─────────────────────────────────────────────────────────────┐
│                      VUI Application                        │
│                  (Running on Host)                          │
└───────────┬─────────────────────────────────────────────────┘
            │
            ├─► Vault (http)            :8200
            ├─► Vault (https)           :8208
            ├─► LDAP Server             :8389
            ├─► LocalStack (AWS Mock)   :4566
            ├─► Keycloak (OIDC/JWT)     :8080
            └─► Kubernetes (kind)       :6443
```

### Core Services

#### 1. **HashiCorp Vault** (`vault`)
- **Purpose**: Main secrets management server
- **Ports**: 8200 (HTTP), 8208 (HTTPS for cert auth)
- **Root Token**: `vui-sandbox-token`
- **Configured Auth Methods**: Token, LDAP, AWS, Kubernetes, OIDC, JWT, UserPass, Certificate

#### 2. **OpenLDAP** (`ldap`)
- **Purpose**: Directory service for LDAP authentication
- **Port**: 8389
- **Domain**: `sandbox.vui`
- **Test User**: `testuser` / `testpassword`
- **Admin**: `cn=admin,dc=sandbox,dc=vui`

#### 3. **LocalStack** (`localstack`)
- **Purpose**: Mock AWS services (STS, IAM) for AWS authentication
- **Port**: 4566
- **Services**: AWS STS and IAM for generating temporary credentials

#### 4. **Keycloak** (`oidc`)
- **Purpose**: Identity provider for OIDC and JWT authentication
- **Port**: 8080
- **Realm**: `vui-sandbox-realm`
- **Test User**: `vui` / `vui`
- **Admin**: `admin` / `admin`

#### 5. **Kubernetes (kind)**
- **Purpose**: Local Kubernetes cluster for Kubernetes auth method
- **Cluster Name**: `vui-sandbox`
- **Service Account**: `vui-sandbox-app-sa`

### Initialization Containers

Each authentication method has a dedicated init container that:
- Enables the auth method in Vault
- Configures the auth backend
- Creates necessary policies and roles
- Verifies the configuration by performing a test login

Init containers use idempotent scripts with status files to prevent re-running completed steps.

---

## Configuration: `configs/vui.yaml`

The `vui.yaml` file defines **8 different Vault profiles**, each demonstrating a different authentication method. This configuration file is the bridge between VUI and the sandbox environment.

### Configuration Structure

```yaml
app:
  theme: "dark"                    # UI theme
  refresh_interval: 30             # Auto-refresh interval (seconds)
  max_secret_size: 10240          # Maximum secret size (bytes)
  clipboard_timeout: 5            # Clipboard clear timeout (seconds)

ui:
  show_hidden_secrets: false      # Show secrets starting with .
  confirm_deletions: true         # Confirm before deleting secrets
  auto_refresh: true              # Enable auto-refresh
  tree_width: 40                  # Width of secrets tree panel

vaults:
  # Profile definitions...
```

### Vault Profiles Explained

Each profile in the `vaults` section represents a **different authentication scenario**:

#### 1. **Token Authentication** (`vault-token`)
```yaml
vault-token:
  address: "http://localhost:8200"
  auth_method: "token"
  auth_config:
    token: "${VAULT_TOKEN}"        # Direct token access
```
- **Use Case**: Simple development, CI/CD, root access
- **Sandbox Credentials**: Token = `vui-sandbox-token`

#### 2. **LDAP Authentication** (`vault-ldap`)
```yaml
vault-ldap:
  address: "http://localhost:8200"
  auth_method: "ldap"
  auth_config:
    username: "${LDAP_USERNAME}"   # testuser
    password: "${LDAP_PASSWORD}"   # testpassword
```
- **Use Case**: Enterprise directory integration (Active Directory, OpenLDAP)
- **Sandbox Server**: OpenLDAP on port 8389

#### 3. **AWS IAM Authentication** (`vault-aws`)
```yaml
vault-aws:
  address: "http://localhost:8200"
  auth_method: "aws"
  auth_config:
    aws_role: "vui-iam-role"
    aws_access_key_id: "${AWS_ACCESS_KEY_ID}"
    aws_secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
    aws_session_token: "${AWS_SESSION_TOKEN}"
```
- **Use Case**: AWS EC2 instances, Lambda functions, ECS tasks
- **Sandbox Mock**: LocalStack simulates AWS STS/IAM

#### 4. **Kubernetes Authentication** (`vault-k8s`)
```yaml
vault-k8s:
  address: "http://localhost:8200"
  auth_method: "kubernetes"
  auth_config:
    k8s_role: "vui-k8s-role"
    k8s_service_account: "vui-sandbox-app-sa"
    k8s_audience: "vault"
    k8s_ttl: 3600
```
- **Use Case**: Pods running in Kubernetes clusters
- **Sandbox Cluster**: kind cluster named `vui-sandbox`

#### 5. **OIDC Authentication** (`vault-oidc`)
```yaml
vault-oidc:
  address: "http://localhost:8200"
  auth_method: "oidc"
  auth_config:
    oidc_role: "vui-oidc-role"
```
- **Use Case**: Interactive login via web browser (SSO)
- **Sandbox Provider**: Keycloak at `http://localhost:8080`
- **Interactive**: Opens browser for OAuth2 flow

#### 6. **JWT Authentication** (`vault-jwt`)
```yaml
vault-jwt:
  address: "http://localhost:8200"
  auth_method: "jwt"
  auth_config:
    jwt_role: "vui-jwt-role"
    jwt: "${JWT}"                  # Pre-obtained JWT token
```
- **Use Case**: Service-to-service auth with JWT tokens
- **Sandbox Provider**: Keycloak-issued JWT tokens

#### 7. **UserPass Authentication** (`vault-userpass`)
```yaml
vault-userpass:
  address: "http://localhost:8200"
  auth_method: "userpass"
  auth_config:
    username: "${USERPASS_USERNAME}"  # vui
    password: "${USERPASS_PASSWORD}"  # vui
```
- **Use Case**: Simple username/password authentication
- **Sandbox Credentials**: `vui` / `vui`

#### 8. **Certificate Authentication** (`vault-cert`)
```yaml
vault-cert:
  address: "https://localhost:8208"  # Note: HTTPS with TLS
  cert_path: "./sandbox/files/vault.crt"
  auth_method: "cert"
  auth_config:
    cert_name: "vui-sandbox-user"
    cert_crt_path: "./sandbox/vol/certs/client.crt"
    cert_key_path: "./sandbox/vol/certs/client.key"
```
- **Use Case**: mTLS authentication for services
- **Sandbox Setup**: Auto-generated client certificates

### Environment Variable Substitution

Notice the `${VARIABLE_NAME}` syntax in the configuration. VUI automatically resolves these from environment variables at runtime. This allows:
- **Security**: Keep sensitive values out of config files
- **Flexibility**: Different credentials per environment
- **CI/CD**: Easy integration with secrets management

---

## Makefile Integration: `env` Target

The `make env` target orchestrates the entire sandbox vui configuration:

```bash
make env
```

### What It Does

It prepares all required environment variables for [VUI profiles](../configs/vui.yaml).

```makefile
env:
    # 1. Export AWS credentials from LocalStack
    export AWS_CONFIG_FILE=./cfg/aws && \
    eval $(aws configure export-credentials --profile vui-iam-role --format env)
    
    # 2. Export LDAP credentials
    export LDAP_USERNAME=testuser && \
    export LDAP_PASSWORD=testpassword
    
    # 3. Export Vault token
    export VAULT_TOKEN=vui-sandbox-token
    
    # 4. Switch kubectl context to sandbox cluster
    kubectl config use-context kind-vui-sandbox
    
    # 5. Obtain JWT token from Keycloak
    export JWT=$(curl -X POST 'http://oidc:8080/realms/vui-sandbox-realm/protocol/openid-connect/token' \
        -H 'Content-Type: application/x-www-form-urlencoded' \
        -d 'client_id=vui-sandbox-oidc-client-id' \
        -d 'client_secret=vui-sandbox-oidc-client-secret' \
        -d 'username=vui' \
        -d 'password=vui' \
        -d 'grant_type=password' \
        -d 'scope=email profile' | jq -r '.access_token')
    
    # 6. Export UserPass credentials
    export USERPASS_USERNAME=vui && \
    export USERPASS_PASSWORD=vui
    
    # 7. Display OIDC credentials for manual login
    echo "OIDC Credentials"
    echo " - Username: vui"
    echo " - Password: vui"
```

The target is integrated with the Makefile in repo root - `make sbx-run`,
which will eval all generated env variables and run VUI.

### Environment Setup Breakdown

| Auth Method | Environment Variables | Source |
|-------------|----------------------|--------|
| **Token** | `VAULT_TOKEN` | Hardcoded dev token |
| **LDAP** | `LDAP_USERNAME`, `LDAP_PASSWORD` | OpenLDAP test user |
| **AWS** | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN` | LocalStack via AWS CLI |
| **Kubernetes** | kubectl context | kind cluster |
| **JWT** | `JWT` | Keycloak token endpoint |
| **UserPass** | `USERPASS_USERNAME`, `USERPASS_PASSWORD` | Vault UserPass auth |
| **OIDC** | Manual login | Browser-based (interactive) |
| **Certificate** | File paths | Auto-generated certs |

### Why This Approach?

1. **Single Command**: Developer runs one command to get a fully configured environment
2. **Realistic Testing**: Tests actual credential flows (AWS CLI, Keycloak tokens, etc.)
3. **No Manual Setup**: All credentials are programmatically obtained
4. **Profile Switching**: VUI can switch between all 8 profiles without restart

---

## Usage Guide

### Initial Setup

```bash
# 1. Add hostname to /etc/hosts (required for Keycloak OIDC)
echo "127.0.0.1    oidc" | sudo tee -a /etc/hosts

# 2. Build the sandbox init image
make build

# 3. Start the sandbox environment (creates kind cluster + docker services)
make up

# 4. Wait for initialization (all init containers should exit with 0)
make ps

# Expected output:
#   init_vault_ldap        Exited (0)
#   init_vault_aws         Exited (0)
#   init_vault_k8s         Exited (0)
#   init_vault_oidc        Exited (0)
#   init_vault_jwt         Exited (0)
#   init_vault_userpass    Exited (0)
#   init_vault_cert        Exited (0)
#   init_localstack        Exited (0)
```

### Running VUI

```bash
# Run VUI with all sandbox profiles configured
make sbx-run  # should be executed from repo root
```

Inside VUI:
1. Press `Tab` to see all 8 vault profiles
2. Select a profile and press `Enter` to authenticate
3. Navigate secrets, create/edit/delete operations
4. Press `Tab` again to switch to a different authentication method

### Troubleshooting

```bash
# View logs from all services
make logs

# Check specific service
make logs vault
make logs localstack

# Restart a specific service
docker-compose restart vault

# Full cleanup and restart
make down
make clean
make up
```

### Cleanup

```bash
make down
```

---

## File Structure

```
sandbox/
├── README.md                      # This file
├── docker-compose.yml             # Service definitions
├── Dockerfile                     # Init container image
├── kind.yml                       # Kubernetes cluster config
├── cfg/
│   └── aws                        # AWS CLI config for LocalStack
├── files/
│   ├── kubernetes.yaml            # K8s resources (SA, RBAC)
│   ├── ldap_bootstrap.ldif        # LDAP initial data
│   ├── oidc_realm_config.json     # Keycloak realm configuration
│   ├── vault_config.hcl           # Vault server config
│   └── vault-san.conf             # Certificate SAN config
├── scripts/
│   ├── init_localstack.sh         # Setup LocalStack AWS resources
│   ├── init_vault_aws.sh          # Configure Vault AWS auth
│   ├── init_vault_cert.sh         # Configure Vault cert auth
│   ├── init_vault_jwt.sh          # Configure Vault JWT auth
│   ├── init_vault_k8s.sh          # Configure Vault K8s auth
│   ├── init_vault_ldap.sh         # Configure Vault LDAP auth
│   ├── init_vault_oidc.sh         # Configure Vault OIDC auth
│   └── init_vault_userpass.sh     # Configure Vault UserPass auth
└── vol/                           # Generated at runtime (gitignored)
    ├── certs/                     # Client certificates
    ├── kind/                      # Kind cluster data
    ├── localstack/                # LocalStack data
    └── vault/                     # Vault data
```

---

## Testing Different Auth Methods

### Example: Testing LDAP Auth

1. Start VUI: `make run` / `make sbx-run`
2. Press `Tab` to open profiles table
3. Navigate to `vault-ldap` and press `Enter`
4. VUI authenticates using `testuser` / `testpassword` from environment
5. Explore secrets in the LDAP-authenticated session

### Example: Testing OIDC Auth

1. Start VUI from repo root: `make sbx-run`
2. Press `Tab` to open profiles table
3. Navigate to `vault-oidc` and press `Enter`
4. Browser opens to Keycloak login page
5. Login with `vui` / `vui`
6. Browser redirects back, VUI is authenticated

### Example: Testing Kubernetes Auth

1. Start VUI from repo root: `make sbx-run`
2. Press `Tab` to open profiles table
3. Navigate to `vault-k8s` and press `Enter`
4. VUI uses service account token from kind cluster
5. Authenticate and access secrets

---

## Advanced Configuration

### Custom Policies

Each auth method is configured with an `rw-policy` that grants full access to `secret/*` paths:

```hcl
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

To test more restrictive policies, modify the init scripts in `sandbox/scripts/`.

### Adding New Auth Methods

To add a new authentication method:

1. Add service to `docker-compose.yml` (if external dependency needed)
2. Create init script in `sandbox/scripts/init_vault_<method>.sh`
3. Add init container to `docker-compose.yml`
4. Add vault profile to `configs/vui.yaml`
5. Update `make dev-run` to export required environment variables

---

## Requirements

### Software Dependencies

- **Docker** or **Podman**: Container runtime
- **kind**: Kubernetes in Docker (`brew install kind`)
- **kubectl**: Kubernetes CLI (`brew install kubectl`)
- **jq**: JSON processor (`brew install jq`)
- **aws-cli**: For LocalStack credentials (`brew install awscli`)

### System Requirements

- **OS**: macOS, Linux (Windows with WSL2)
- **RAM**: 4GB minimum (8GB recommended)
- **Disk**: 2GB free space
- **Network**: Ports 8200, 8208, 8389, 4566, 8080, 6443 available

---

## FAQ

**Q: Why does `make sbx-run` fail with "connection refused"?**  
A: Ensure all init containers have completed successfully (`make sbx-ps`). If any failed, check logs (`make sbx-logs`) and restart.

**Q: Can I run VUI against only one auth method?**  
A: Yes! Comment out unwanted profiles in `configs/vui.yaml` or start only specific services in docker-compose.

**Q: How do I reset the environment?**  
A: Run `make down` then `make clean` to remove all generated files, then `make up` to start fresh.

**Q: Can I use this sandbox for production?**  
A: **NO!** This is a development environment with hardcoded credentials and insecure settings. Never expose these services to the internet.

**Q: The OIDC auth doesn't work**  
A: Ensure you added `127.0.0.1 oidc` to `/etc/hosts`. Keycloak needs this hostname for proper redirect URIs.

---

## Resources

- [HashiCorp Vault Documentation](https://www.vaultproject.io/docs)
- [Vault Auth Methods](https://www.vaultproject.io/docs/auth)
- [kind Documentation](https://kind.sigs.k8s.io/)
- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [LocalStack Documentation](https://docs.localstack.cloud/)
- [OpenLDAP](https://www.openldap.org/)
