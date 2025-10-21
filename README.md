# VUI (Vault UI)

A Console User Interface (CUI) application for HashiCorp Vault, inspired by derailed/k9s.

## Overview

VUI provides an intuitive terminal-based interface for exploring and managing secrets in HashiCorp Vault. The application supports multiple vault connections, hierarchical secret navigation, and full CRUD operations.

### 🔧 Dependencies

- **Vault API**: `github.com/hashicorp/vault/api` - Official HashiCorp Vault client
- **Configuration**: `github.com/spf13/viper` - Configuration management
- **Terminal UI**: `github.com/rivo/tview` - Terminal user interface framework
- **Terminal Control**: `github.com/gdamore/tcell/v2` - Terminal control library
- **Clipboard**: `github.com/atotto/clipboard` - Cross-platform clipboard access
- **Logging**: `github.com/sirupsen/logrus` - Structured logging
- **Testing**: `github.com/stretchr/testify` - Test assertions

## Installation

### Prerequisites

- Go 1.21 or later
- HashiCorp Vault server (optional - application starts gracefully without it)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/rvolykh/vui.git
cd vui

# Build the application
make build

# Run the application
./vui
```

## Usage

### Vault Profiles Screen

When no vault servers are connected, you'll see:

- **Welcome message** with connection status
- **Navigation instructions** and keyboard shortcuts
- **List of configured vault profiles** with their connection status:
  - ✅ **Connected**: Vault is reachable and unsealed
  - 🔒 **Sealed**: Vault is reachable but sealed
  - ❌ **Disconnected**: Vault is not reachable

### Keyboard Shortcuts

#### Navigation
- `↑/↓`: Navigate tree items
- `←/→`: Collapse/expand tree nodes
- `Enter`: Select item or enter directory
- `Esc`: Go back or cancel
- `Tab`: Navigate form fields (in forms)

#### Secret Panel
- `c`: Create new secret
- `e`: Edit selected secret
- `Ctrl+d`: Delete selected secret
- `d`: Unmask/mask secret value
- `v`: Copy secret value to clipboard

#### Vault Management
- `Tab`: Switch vault profiles (shows profiles table)
- `Esc`: Go back to secrets (if previously selected a profile)

#### Global
- `h/F1`: Show help
- `r/F5`: Refresh
- `q/Ctrl+C`: Exit application

### Available Make Targets

```bash
make build        # Build the application
make test         # Run tests
make clean        # Clean build artifacts
make deps         # Download dependencies
make fmt          # Format code
make vet          # Run go vet
make help         # Show all available targets
```

## Configuration

VUI uses YAML configuration files with environment variable support.

### Default Configuration

The application looks for configuration in:
1. `./configs/vui.yaml`
2. `$HOME/.vui/vui.yaml`
3. `/etc/vui/vui.yaml`

### Example Configuration

```yaml
app:
  theme: "dark"
  refresh_interval: 30

ui:
  show_hidden_secrets: false
  confirm_deletions: true
  auto_refresh: true

vaults:
  local:
    address: "http://localhost:8200"
    auth_method: "token"
    token: "${VAULT_TOKEN}" # variable will be read from environment variables once app is started
    namespace: ""
```

### Advanced Authentication Examples

#### LDAP Authentication
```yaml
vaults:
  ldap_vault:
    address: "https://vault.company.com"
    auth_method: "ldap"
    namespace: "production"
    auth_config:
      username: "${VAULT_USERNAME}"
      password: "${VAULT_PASSWORD}"
```

#### AWS IAM Authentication
```yaml
vaults:
  aws_vault:
    address: "https://vault.company.com"
    auth_method: "aws"
    namespace: "aws"
    auth_config:
      aws_access_key_id: "${AWS_ACCESS_KEY_ID}"
      aws_secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
      aws_role: "vault-role"
      aws_region: "us-east-1"
```

#### Azure Authentication
```yaml
vaults:
  azure_vault:
    address: "https://vault.company.com"
    auth_method: "azure"
    namespace: "azure"
    auth_config:
      azure_tenant_id: "${AZURE_TENANT_ID}"
      azure_client_id: "${AZURE_CLIENT_ID}"
      azure_client_secret: "${AZURE_CLIENT_SECRET}"
      azure_role: "vault-role"
```

#### Kubernetes Authentication
```yaml
vaults:
  k8s_vault:
    address: "https://vault.company.com"
    auth_method: "kubernetes"
    namespace: "k8s"
    auth_config:
      k8s_role: "vault-role"
      k8s_token_path: "/var/run/secrets/kubernetes.io/serviceaccount/token"
```

#### JWT Authentication
```yaml
vaults:
  jwt_vault:
    address: "https://vault.company.com"
    auth_method: "jwt"
    namespace: "jwt"
    auth_config:
      jwt_role: "vault-role"
      jwt: "${JWT_TOKEN}"
```

#### Certificate Authentication
```yaml
vaults:
  cert_vault:
    address: "https://vault.company.com"
    auth_method: "cert"
    namespace: "cert"
    auth_config:
      cert_name: "vault-client"
      cert_path: "/path/to/client.crt"
      key_path: "/path/to/client.key"
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific package tests
go test ./internal/config/...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make vet

# Run comprehensive linter (requires golangci-lint)
make lint
```

## Sandbox

Local playground environment with different vault auths / profiles.

### Requirements

- [kind](https://kind.sigs.k8s.io/)
- [docker](https://www.docker.com/) / [podman](https://podman.io/)
- `127.0.0.1    oidc` record on `/etc/hosts`

### Commands

```bash
make dev-build
      # Build Sandbox environment init image

make dev-up
      # Start Sandbox environment

make dev-ps
      # Check Sabdbox environment status
      # All `init_` containers has to have status "Exited (0)"

make dev-logs
      # Check logs of Sandbox environment

make dev-run
      # Run VUI with necessary environment variables

make dev-down
      # Stop Sandbox environment

make clean
      # Cleanup temporary files, required when Sandbox should be re-created
```

## Acknowledgments

- Inspired by [derailed/k9s](https://github.com/derailed/k9s) for Kubernetes management
- Built with [HashiCorp Vault](https://www.vaultproject.io/) API
- Uses [tview](https://github.com/rivo/tview) and [tcell](https://github.com/gdamore/tcell) for terminal UI
