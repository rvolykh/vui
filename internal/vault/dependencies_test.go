package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
)

// MockVaultAuthenticator is a mock implementation of vaultAuthenticator
type MockVaultAuthenticator struct {
	VerifyAuthenticationFunc func(client *api.Client) error
	AuthenticateFunc         func(client *api.Client, profile *config.VaultProfile) error

	// Call tracking
	VerifyAuthenticationCalls int
	AuthenticateCalls         int
	LastClient                *api.Client
	LastProfile               *config.VaultProfile
}

func NewMockVaultAuthenticator() *MockVaultAuthenticator {
	return &MockVaultAuthenticator{}
}

func (m *MockVaultAuthenticator) VerifyAuthentication(client *api.Client) error {
	m.VerifyAuthenticationCalls++
	m.LastClient = client

	if m.VerifyAuthenticationFunc != nil {
		return m.VerifyAuthenticationFunc(client)
	}
	return nil
}

func (m *MockVaultAuthenticator) Authenticate(client *api.Client, profile *config.VaultProfile) error {
	m.AuthenticateCalls++
	m.LastClient = client
	m.LastProfile = profile

	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(client, profile)
	}
	return nil
}

// Reset resets the call tracking
func (m *MockVaultAuthenticator) Reset() {
	m.VerifyAuthenticationCalls = 0
	m.AuthenticateCalls = 0
	m.LastClient = nil
	m.LastProfile = nil
}

// MockVaultSecretsManager is a mock implementation of vaultSecretsManager
type MockVaultSecretsManager struct {
	ListSecretsFunc          func(path string) ([]*SecretNode, error)
	GetSecretFunc            func(path string) (*SecretNode, error)
	CreateSecretFunc         func(path string, data map[string]interface{}) error
	UpdateSecretFunc         func(path string, data map[string]interface{}) error
	DeleteSecretFunc         func(path string) error
	BuildTreeFunc            func(rootPath string, maxDepth int) (*SecretNode, error)
	SearchSecretsFunc        func(pattern string, rootPath string) ([]*SecretNode, error)
	SearchSecretsByValueFunc func(valuePattern string, rootPath string) ([]*SearchResult, error)
	SearchSecretsByKeyFunc   func(keyPattern string, rootPath string) ([]*SearchResult, error)

	// Call tracking
	ListSecretsCalls          int
	GetSecretCalls            int
	CreateSecretCalls         int
	UpdateSecretCalls         int
	DeleteSecretCalls         int
	BuildTreeCalls            int
	SearchSecretsCalls        int
	SearchSecretsByValueCalls int
	SearchSecretsByKeyCalls   int

	// Last call parameters
	LastPath         string
	LastData         map[string]interface{}
	LastRootPath     string
	LastMaxDepth     int
	LastPattern      string
	LastValuePattern string
	LastKeyPattern   string
}

func NewMockVaultSecretsManager() *MockVaultSecretsManager {
	return &MockVaultSecretsManager{}
}

func (m *MockVaultSecretsManager) ListSecrets(path string) ([]*SecretNode, error) {
	m.ListSecretsCalls++
	m.LastPath = path

	if m.ListSecretsFunc != nil {
		return m.ListSecretsFunc(path)
	}
	return []*SecretNode{}, nil
}

func (m *MockVaultSecretsManager) GetSecret(path string) (*SecretNode, error) {
	m.GetSecretCalls++
	m.LastPath = path

	if m.GetSecretFunc != nil {
		return m.GetSecretFunc(path)
	}
	return &SecretNode{
		Name:     "test-secret",
		Path:     path,
		IsSecret: true,
		Data:     map[string]interface{}{},
	}, nil
}

func (m *MockVaultSecretsManager) CreateSecret(path string, data map[string]interface{}) error {
	m.CreateSecretCalls++
	m.LastPath = path
	m.LastData = data

	if m.CreateSecretFunc != nil {
		return m.CreateSecretFunc(path, data)
	}
	return nil
}

func (m *MockVaultSecretsManager) UpdateSecret(path string, data map[string]interface{}) error {
	m.UpdateSecretCalls++
	m.LastPath = path
	m.LastData = data

	if m.UpdateSecretFunc != nil {
		return m.UpdateSecretFunc(path, data)
	}
	return nil
}

func (m *MockVaultSecretsManager) DeleteSecret(path string) error {
	m.DeleteSecretCalls++
	m.LastPath = path

	if m.DeleteSecretFunc != nil {
		return m.DeleteSecretFunc(path)
	}
	return nil
}

func (m *MockVaultSecretsManager) BuildTree(rootPath string, maxDepth int) (*SecretNode, error) {
	m.BuildTreeCalls++
	m.LastRootPath = rootPath
	m.LastMaxDepth = maxDepth

	if m.BuildTreeFunc != nil {
		return m.BuildTreeFunc(rootPath, maxDepth)
	}
	return &SecretNode{
		Name:     "root",
		Path:     rootPath,
		IsSecret: false,
		Children: []*SecretNode{},
	}, nil
}

func (m *MockVaultSecretsManager) SearchSecrets(pattern string, rootPath string) ([]*SecretNode, error) {
	m.SearchSecretsCalls++
	m.LastPattern = pattern
	m.LastRootPath = rootPath

	if m.SearchSecretsFunc != nil {
		return m.SearchSecretsFunc(pattern, rootPath)
	}
	return []*SecretNode{}, nil
}

func (m *MockVaultSecretsManager) SearchSecretsByValue(valuePattern string, rootPath string) ([]*SearchResult, error) {
	m.SearchSecretsByValueCalls++
	m.LastValuePattern = valuePattern
	m.LastRootPath = rootPath

	if m.SearchSecretsByValueFunc != nil {
		return m.SearchSecretsByValueFunc(valuePattern, rootPath)
	}
	return []*SearchResult{}, nil
}

func (m *MockVaultSecretsManager) SearchSecretsByKey(keyPattern string, rootPath string) ([]*SearchResult, error) {
	m.SearchSecretsByKeyCalls++
	m.LastKeyPattern = keyPattern
	m.LastRootPath = rootPath

	if m.SearchSecretsByKeyFunc != nil {
		return m.SearchSecretsByKeyFunc(keyPattern, rootPath)
	}
	return []*SearchResult{}, nil
}

// Reset resets the call tracking
func (m *MockVaultSecretsManager) Reset() {
	m.ListSecretsCalls = 0
	m.GetSecretCalls = 0
	m.CreateSecretCalls = 0
	m.UpdateSecretCalls = 0
	m.DeleteSecretCalls = 0
	m.BuildTreeCalls = 0
	m.SearchSecretsCalls = 0
	m.SearchSecretsByValueCalls = 0
	m.SearchSecretsByKeyCalls = 0

	m.LastPath = ""
	m.LastData = nil
	m.LastRootPath = ""
	m.LastMaxDepth = 0
	m.LastPattern = ""
	m.LastValuePattern = ""
	m.LastKeyPattern = ""
}

// MockVaultConnectionManager is a mock implementation of vaultConnectionManager
type MockVaultConnectionManager struct {
	AddConnectionFunc           func(name string, client *Client)
	TestConnectionAsyncFunc     func(name string)
	RemoveConnectionFunc        func(name string)
	GetConnectionFunc           func(name string) (*Client, error)
	GetConnectionStatusFunc     func(name string) (*ConnectionStatus, error)
	ListConnectionsFunc         func() []string
	RefreshConnectionStatusFunc func(name string) error
	RefreshAllConnectionsFunc   func()
	SetAllConnectingFunc        func()
	GetHealthyConnectionsFunc   func() []string
	GetConnectedConnectionsFunc func() []string

	// Call tracking
	AddConnectionCalls           int
	TestConnectionAsyncCalls     int
	RemoveConnectionCalls        int
	GetConnectionCalls           int
	GetConnectionStatusCalls     int
	ListConnectionsCalls         int
	RefreshConnectionStatusCalls int
	RefreshAllConnectionsCalls   int
	SetAllConnectingCalls        int
	GetHealthyConnectionsCalls   int
	GetConnectedConnectionsCalls int

	// Last call parameters
	LastName   string
	LastClient *Client

	// Mock data
	Connections        map[string]*Client
	ConnectionStatuses map[string]*ConnectionStatus
}

func NewMockVaultConnectionManager() *MockVaultConnectionManager {
	return &MockVaultConnectionManager{
		Connections:        make(map[string]*Client),
		ConnectionStatuses: make(map[string]*ConnectionStatus),
	}
}

func (m *MockVaultConnectionManager) AddConnection(name string, client *Client) {
	m.AddConnectionCalls++
	m.LastName = name
	m.LastClient = client

	if m.AddConnectionFunc != nil {
		m.AddConnectionFunc(name, client)
		return
	}

	m.Connections[name] = client
}

func (m *MockVaultConnectionManager) TestConnectionAsync(name string) {
	m.TestConnectionAsyncCalls++
	m.LastName = name

	if m.TestConnectionAsyncFunc != nil {
		m.TestConnectionAsyncFunc(name)
	}
}

func (m *MockVaultConnectionManager) RemoveConnection(name string) {
	m.RemoveConnectionCalls++
	m.LastName = name

	if m.RemoveConnectionFunc != nil {
		m.RemoveConnectionFunc(name)
		return
	}

	delete(m.Connections, name)
	delete(m.ConnectionStatuses, name)
}

func (m *MockVaultConnectionManager) GetConnection(name string) (*Client, error) {
	m.GetConnectionCalls++
	m.LastName = name

	if m.GetConnectionFunc != nil {
		return m.GetConnectionFunc(name)
	}

	if client, ok := m.Connections[name]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("connection '%s' not found", name)
}

func (m *MockVaultConnectionManager) GetConnectionStatus(name string) (*ConnectionStatus, error) {
	m.GetConnectionStatusCalls++
	m.LastName = name

	if m.GetConnectionStatusFunc != nil {
		return m.GetConnectionStatusFunc(name)
	}

	if status, ok := m.ConnectionStatuses[name]; ok {
		return status, nil
	}
	return &ConnectionStatus{
		Address:    "http://localhost:8200",
		Connected:  false,
		Connecting: false,
	}, nil
}

func (m *MockVaultConnectionManager) ListConnections() []string {
	m.ListConnectionsCalls++

	if m.ListConnectionsFunc != nil {
		return m.ListConnectionsFunc()
	}

	names := make([]string, 0, len(m.Connections))
	for name := range m.Connections {
		names = append(names, name)
	}
	return names
}

func (m *MockVaultConnectionManager) RefreshConnectionStatus(name string) error {
	m.RefreshConnectionStatusCalls++
	m.LastName = name

	if m.RefreshConnectionStatusFunc != nil {
		return m.RefreshConnectionStatusFunc(name)
	}
	return nil
}

func (m *MockVaultConnectionManager) RefreshAllConnections() {
	m.RefreshAllConnectionsCalls++

	if m.RefreshAllConnectionsFunc != nil {
		m.RefreshAllConnectionsFunc()
	}
}

func (m *MockVaultConnectionManager) SetAllConnecting() {
	m.SetAllConnectingCalls++

	if m.SetAllConnectingFunc != nil {
		m.SetAllConnectingFunc()
		return
	}

	for name := range m.ConnectionStatuses {
		if m.ConnectionStatuses[name] != nil {
			m.ConnectionStatuses[name].Connecting = true
			m.ConnectionStatuses[name].Connected = false
		}
	}
}

func (m *MockVaultConnectionManager) GetHealthyConnections() []string {
	m.GetHealthyConnectionsCalls++

	if m.GetHealthyConnectionsFunc != nil {
		return m.GetHealthyConnectionsFunc()
	}

	healthy := make([]string, 0)
	for name, status := range m.ConnectionStatuses {
		if status != nil && status.Connected && !status.Sealed {
			healthy = append(healthy, name)
		}
	}
	return healthy
}

func (m *MockVaultConnectionManager) GetConnectedConnections() []string {
	m.GetConnectedConnectionsCalls++

	if m.GetConnectedConnectionsFunc != nil {
		return m.GetConnectedConnectionsFunc()
	}

	connected := make([]string, 0)
	for name, status := range m.ConnectionStatuses {
		if status != nil && status.Connected {
			connected = append(connected, name)
		}
	}
	return connected
}

// Reset resets the call tracking
func (m *MockVaultConnectionManager) Reset() {
	m.AddConnectionCalls = 0
	m.TestConnectionAsyncCalls = 0
	m.RemoveConnectionCalls = 0
	m.GetConnectionCalls = 0
	m.GetConnectionStatusCalls = 0
	m.ListConnectionsCalls = 0
	m.RefreshConnectionStatusCalls = 0
	m.RefreshAllConnectionsCalls = 0
	m.SetAllConnectingCalls = 0
	m.GetHealthyConnectionsCalls = 0
	m.GetConnectedConnectionsCalls = 0

	m.LastName = ""
	m.LastClient = nil
}

// Helper methods for setting up mock data

func (m *MockVaultConnectionManager) SetupConnection(name string, connected bool, sealed bool) {
	m.Connections[name] = &Client{}
	m.ConnectionStatuses[name] = &ConnectionStatus{
		Address:    fmt.Sprintf("http://%s.vault.local:8200", name),
		Connected:  connected,
		Connecting: false,
		Sealed:     sealed,
		Error:      "",
	}
}

func (m *MockVaultConnectionManager) SetupConnectionWithError(name string, errMsg string) {
	m.Connections[name] = &Client{}
	m.ConnectionStatuses[name] = &ConnectionStatus{
		Address:    fmt.Sprintf("http://%s.vault.local:8200", name),
		Connected:  false,
		Connecting: false,
		Sealed:     false,
		Error:      errMsg,
	}
}

// MockDialogService is a mock implementation for dialog services
type MockDialogService struct {
	ShowInfoCalls  int
	ShowErrorCalls int
	LastTitle      string
	LastMessage    string
}

func NewMockDialogService() *MockDialogService {
	return &MockDialogService{}
}

func (m *MockDialogService) ShowInfo(title, message string, callback func()) {
	m.ShowInfoCalls++
	m.LastTitle = title
	m.LastMessage = message
	if callback != nil {
		callback()
	}
}

func (m *MockDialogService) ShowError(message string, callback func()) {
	m.ShowErrorCalls++
	m.LastMessage = message
	if callback != nil {
		callback()
	}
}

func (m *MockDialogService) Reset() {
	m.ShowInfoCalls = 0
	m.ShowErrorCalls = 0
	m.LastTitle = ""
	m.LastMessage = ""
}
