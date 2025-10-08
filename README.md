# VUI (Vault UI)

A Console User Interface (CUI) application for HashiCorp Vault, inspired by derailed/k9s.

## Overview

VUI provides an intuitive terminal-based interface for exploring and managing secrets in HashiCorp Vault. The application supports multiple vault connections, hierarchical secret navigation, and full CRUD operations.

### ✅ Completed Features

#### Phase 1 - Foundation ✅
- **Project Structure**: Clean, modular Go project structure
- **Configuration Management**: YAML-based configuration with environment variable support
- **Vault Client Integration**: Basic HashiCorp Vault API client with connection management
- **Multi-Vault Support**: Framework for managing multiple vault connections
- **Build System**: Makefile with comprehensive build targets
- **Testing Framework**: Basic test structure with configuration tests

#### Phase 2 - Core Vault Operations ✅
- **Connection Management**: Advanced vault connection monitoring with status tracking
- **Secrets Management**: Complete CRUD operations for vault secrets
- **Hierarchical Navigation**: Tree-based secret structure with recursive listing
- **Profile Management**: Vault connection profiles with validation
- **Error Handling**: Comprehensive error handling and connection status reporting
- **Search Functionality**: Basic secret search capabilities
- **Metadata Support**: Secret versioning and metadata tracking

#### Phase 3 - User Interface ✅
- **Terminal UI Framework**: Full tview-based terminal interface
- **Main Layout**: Three-panel layout with tree, secret view, and status bar
- **Directory Tree**: Interactive tree component for secret navigation
- **Secret Display**: Rich secret viewing with metadata and formatting
- **Keyboard Navigation**: Intuitive keyboard shortcuts and navigation
- **Status Bar**: Real-time vault connection status and selection info
- **Input Forms**: Forms for secret creation, editing, and deletion
- **Help System**: Built-in help with keyboard shortcut reference
- **Graceful Startup**: Starts even without vault server, shows connection status
- **Vault Profiles**: Displays all configured vault profiles with connection status
- **Offline Mode**: Works in offline mode with helpful error messages
- **Connection Error Handling**: Graceful handling of connection failures

#### Phase 4 - Secret Management ✅
- **Form Integration**: Complete integration of forms with main UI using modal system
- **Modal System**: Seamless modal dialogs for forms and confirmations
- **Clipboard Integration**: Full clipboard support for secret values and metadata
- **Copy Functionality**: Copy entire secrets or individual values to system clipboard
- **Value Selection**: Smart value selection for multi-key secrets
- **Success Feedback**: Visual feedback for clipboard operations
- **Enhanced Keyboard Shortcuts**: Updated shortcuts for all form and clipboard operations

#### Phase 5 - Advanced Features ✅
- **Enhanced Multi-Vault Support**: Advanced vault connection management with profile-based configuration
- **Vault Profile Management**: Create, edit, and delete vault connection profiles
- **Advanced Search**: Multi-criteria search with support for name, path, key, value, and metadata searches
- **Search Results Display**: Rich search results with relevance scoring and match highlighting
- **Advanced Authentication**: Support for multiple authentication methods including:
  - Token authentication
  - LDAP authentication
  - AWS IAM authentication
  - Azure authentication
  - GCP authentication
  - Kubernetes authentication
  - JWT authentication
  - Userpass authentication
  - Certificate authentication
- **Authentication Validation**: Comprehensive validation for all authentication methods
- **Connection Status Monitoring**: Real-time monitoring of vault connection health
- **Vault Switching**: Seamless switching between multiple vault connections

### 🏗️ Architecture

```
vui/
├── cmd/vui/           # Application entry point
├── internal/
│   ├── app/           # Core application logic
│   ├── config/        # Configuration management
│   │   ├── config.go           # Main configuration
│   │   └── vault_profiles.go   # Vault profiles management
│   ├── vault/         # Vault client and operations
│   │   ├── manager.go          # Vault connection manager
│   │   ├── client.go           # Vault API operations
│   │   ├── connection.go       # Connection status monitoring
│   │   └── secrets.go          # Secrets management
│   └── ui/            # User interface
│       ├── app.go              # Main UI application
│       ├── layout.go           # Application layout
│       ├── tree.go             # Directory tree component
│       ├── secret_view.go      # Secret display component
│       ├── status_bar.go       # Status bar component
│       ├── forms.go            # Input forms
│       └── vault_profiles.go   # Vault profiles display
├── configs/           # Configuration files
│   ├── default.yaml   # Default application config
│   └── vaults.yaml    # Vault connection profiles
└── Makefile          # Build automation
```

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

### Graceful Startup

VUI is designed to start gracefully even when no Vault server is available. When you run the application:

1. **If no Vault servers are connected**: The application shows a welcome screen with all configured vault profiles and their connection status
2. **If Vault servers are available**: The application directly shows the main interface with the secrets tree

### Vault Profiles Screen

When no vault servers are connected, you'll see:

- **Welcome message** with connection status
- **List of configured vault profiles** with their connection status:
  - ✅ **Connected**: Vault is reachable and unsealed
  - 🔒 **Sealed**: Vault is reachable but sealed
  - ❌ **Disconnected**: Vault is not reachable
- **Navigation instructions** and keyboard shortcuts

### Offline Mode

If the application starts but no vault connections are available, it will display:

- **Offline mode message** explaining the situation
- **Troubleshooting steps** to help resolve connection issues
- **Configuration information** showing current vault settings
- **Clear instructions** on how to connect to a vault server

### Keyboard Shortcuts

#### Vault Profiles Screen
- `↑/↓`: Navigate vault profiles
- `Enter`: Connect to selected vault
- `r`: Refresh connection status
- `n`: Add new vault profile
- `F1`: Show help
- `q`: Quit application

#### Main Interface
- `↑/↓`: Navigate tree items
- `←/→`: Collapse/expand tree nodes
- `Enter`: Select item or enter directory
- `Tab`: Switch between panels
- `c`: Create new secret
- `e`: Edit selected secret
- `Ctrl+d`: Delete selected secret
- `r`: Refresh current view
- `s`: Search secrets
- `Ctrl+v`: Switch vault
- `F1`: Show help
- `Ctrl+C`: Exit application

#### Secret Panel
- `c`: Copy entire secret to clipboard
- `v`: Copy secret value to clipboard
- `e`: Edit selected secret

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
1. `./configs/default.yaml`
2. `$HOME/.vui/config.yaml`
3. `/etc/vui/config.yaml`

### Environment Variables

- `VAULT_ADDR` - Vault server address
- `VAULT_TOKEN` - Vault authentication token
- `VAULT_NAMESPACE` - Vault namespace

### Example Configuration

```yaml
app:
  default_vault: "default"
  theme: "dark"
  refresh_interval: 30

vault:
  address: "http://localhost:8200"
  auth_method: "token"
  token: ""
  namespace: ""

ui:
  show_hidden_secrets: false
  confirm_deletions: true
  auto_refresh: true
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
make test-coverage

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

## Roadmap

### Phase 5: Advanced Features ✅
- [x] Enhanced multi-vault support and switching
- [x] Advanced search functionality
- [x] Advanced authentication methods

### Phase 6: Future Enhancements
- [ ] Performance optimizations
- [ ] Additional authentication methods
- [ ] Plugin system for custom integrations
- [ ] Advanced secret management features

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [derailed/k9s](https://github.com/derailed/k9s) for Kubernetes management
- Built with [HashiCorp Vault](https://www.vaultproject.io/) API
- Uses [tview](https://github.com/rivo/tview) for terminal UI (planned)
