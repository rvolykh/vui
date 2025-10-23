package vault

import (
	"fmt"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
)

// Helper function to create a test manager with mocks
func createTestManager(t *testing.T) (*Manager, *MockVaultConnectionManager, *MockVaultAuthenticator, *MockVaultSecretsManager) {
	t.Helper()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	cfg := &config.Config{
		Vaults: map[string]config.VaultProfile{
			"vault1": {
				Address:    "http://vault1.local:8200",
				AuthMethod: "token",
				Namespace:  "ns1",
				AuthConfig: config.AuthConfig{
					Token: "token1",
				},
			},
			"vault2": {
				Address:    "http://vault2.local:8200",
				AuthMethod: "userpass",
				Namespace:  "ns2",
				AuthConfig: config.AuthConfig{
					Username: "user",
					Password: "pass",
				},
			},
		},
	}

	mockConnMgr := NewMockVaultConnectionManager()
	mockAuthMgr := NewMockVaultAuthenticator()
	mockSecretsMgr := NewMockVaultSecretsManager()

	manager := &Manager{
		config:        cfg,
		clients:       make(map[string]*Client),
		connectionMgr: mockConnMgr,
		authMgr:       mockAuthMgr,
		secretsMgr:    mockSecretsMgr,
		logger:        logger,
	}

	return manager, mockConnMgr, mockAuthMgr, mockSecretsMgr
}

// Helper function to create a test client
func createTestClient(t *testing.T, profile *config.VaultProfile) *Client {
	t.Helper()

	apiConfig := api.DefaultConfig()
	apiConfig.Address = profile.Address

	apiClient, err := api.NewClient(apiConfig)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	if profile.Namespace != "" {
		apiClient.SetNamespace(profile.Namespace)
	}

	return &Client{
		apiClient: apiClient,
		profile:   profile,
		logger:    logrus.New(),
	}
}

func TestNewManager(t *testing.T) {
	t.Run("creates manager with valid config", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.Config{
			Vaults: map[string]config.VaultProfile{
				"test-vault": {
					Address:    "http://localhost:8200",
					AuthMethod: "token",
					AuthConfig: config.AuthConfig{
						Token: "test-token",
					},
				},
			},
		}

		manager, err := NewManager(cfg, logger)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if manager == nil {
			t.Fatal("Expected manager to be created")
		}

		if len(manager.clients) != 1 {
			t.Errorf("Expected 1 client, got %d", len(manager.clients))
		}
	})

	t.Run("creates manager with nil logger", func(t *testing.T) {
		cfg := &config.Config{
			Vaults: map[string]config.VaultProfile{
				"test-vault": {
					Address:    "http://localhost:8200",
					AuthMethod: "token",
					AuthConfig: config.AuthConfig{
						Token: "test-token",
					},
				},
			},
		}

		manager, err := NewManager(cfg, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if manager == nil {
			t.Fatal("Expected manager to be created")
		}

		if manager.logger == nil {
			t.Error("Expected logger to be created when nil is passed")
		}
	})

	t.Run("creates manager with empty vaults", func(t *testing.T) {
		logger := logrus.New()
		cfg := &config.Config{
			Vaults: map[string]config.VaultProfile{},
		}

		manager, err := NewManager(cfg, logger)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if manager == nil {
			t.Fatal("Expected manager to be created")
		}

		if len(manager.clients) != 0 {
			t.Errorf("Expected 0 clients, got %d", len(manager.clients))
		}
	})
}

func TestManager_GetActiveClient(t *testing.T) {
	t.Run("returns active client when set", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile := manager.config.Vaults["vault1"]
		client := createTestClient(t, &profile)
		manager.clients["vault1"] = client
		manager.activeVault = "vault1"

		activeClient, err := manager.GetActiveClient()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if activeClient != client {
			t.Error("Expected to get the active client")
		}
	})

	t.Run("returns error when no active vault set", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)
		manager.activeVault = ""

		_, err := manager.GetActiveClient()
		if err == nil {
			t.Error("Expected error when no active vault is set")
		}
	})

	t.Run("returns error when active vault not found", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)
		manager.activeVault = "non-existent"

		_, err := manager.GetActiveClient()
		if err == nil {
			t.Error("Expected error when active vault not found")
		}
	})
}

func TestManager_GetClient(t *testing.T) {
	t.Run("returns client by name", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile := manager.config.Vaults["vault1"]
		client := createTestClient(t, &profile)
		manager.clients["vault1"] = client

		retrievedClient, err := manager.GetClient("vault1")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if retrievedClient != client {
			t.Error("Expected to get the correct client")
		}
	})

	t.Run("returns error when client not found", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		_, err := manager.GetClient("non-existent")
		if err == nil {
			t.Error("Expected error when client not found")
		}
	})
}

func TestManager_SwitchVault(t *testing.T) {
	t.Run("switches to existing vault successfully", func(t *testing.T) {
		manager, _, mockAuthMgr, _ := createTestManager(t)

		profile1 := manager.config.Vaults["vault1"]
		profile2 := manager.config.Vaults["vault2"]
		client1 := createTestClient(t, &profile1)
		client2 := createTestClient(t, &profile2)
		manager.clients["vault1"] = client1
		manager.clients["vault2"] = client2
		manager.activeVault = "vault1"

		// Set up mock authenticator to succeed
		mockAuthMgr.AuthenticateFunc = func(client *api.Client, profile *config.VaultProfile) error {
			return nil
		}
		mockAuthMgr.VerifyAuthenticationFunc = func(client *api.Client) error {
			return nil
		}

		err := manager.SwitchVault("vault2")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if manager.activeVault != "vault2" {
			t.Errorf("Expected active vault to be 'vault2', got '%s'", manager.activeVault)
		}

		if mockAuthMgr.AuthenticateCalls != 1 {
			t.Errorf("Expected Authenticate to be called once, got %d calls", mockAuthMgr.AuthenticateCalls)
		}

		if mockAuthMgr.VerifyAuthenticationCalls != 1 {
			t.Errorf("Expected VerifyAuthentication to be called once, got %d calls", mockAuthMgr.VerifyAuthenticationCalls)
		}
	})

	t.Run("returns error when vault not found", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		err := manager.SwitchVault("non-existent")
		if err == nil {
			t.Error("Expected error when vault not found")
		}
	})

	t.Run("returns error when authentication fails", func(t *testing.T) {
		manager, _, mockAuthMgr, _ := createTestManager(t)

		profile := manager.config.Vaults["vault1"]
		client := createTestClient(t, &profile)
		manager.clients["vault1"] = client

		// Set up mock authenticator to fail
		mockAuthMgr.AuthenticateFunc = func(client *api.Client, profile *config.VaultProfile) error {
			return fmt.Errorf("authentication failed")
		}

		err := manager.SwitchVault("vault1")
		if err == nil {
			t.Error("Expected error when authentication fails")
		}
	})

	t.Run("returns error when authentication verification fails", func(t *testing.T) {
		manager, _, mockAuthMgr, _ := createTestManager(t)

		profile := manager.config.Vaults["vault1"]
		client := createTestClient(t, &profile)
		manager.clients["vault1"] = client

		// Set up mock authenticator: authenticate succeeds but verification fails
		mockAuthMgr.AuthenticateFunc = func(client *api.Client, profile *config.VaultProfile) error {
			return nil
		}
		mockAuthMgr.VerifyAuthenticationFunc = func(client *api.Client) error {
			return fmt.Errorf("verification failed")
		}

		err := manager.SwitchVault("vault1")
		if err == nil {
			t.Error("Expected error when authentication verification fails")
		}
	})

	t.Run("returns error when profile not found", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile := manager.config.Vaults["vault1"]
		client := createTestClient(t, &profile)
		manager.clients["vault1"] = client

		// Remove profile from config but keep client
		delete(manager.config.Vaults, "vault1")

		err := manager.SwitchVault("vault1")
		if err == nil {
			t.Error("Expected error when profile not found")
		}
	})
}

func TestManager_AddVault(t *testing.T) {
	t.Run("returns error when vault already exists", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile := &config.VaultProfile{
			Address:    "http://duplicate.local:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{
				Token: "token",
			},
		}

		err := manager.AddVault("vault1", profile)
		if err == nil {
			t.Error("Expected error when vault already exists")
		}
	})

	// Note: Testing successful AddVault requires file I/O for config.Save()
	// which is not suitable for unit tests. Integration tests should cover this.
}

func TestManager_ListVaults(t *testing.T) {
	t.Run("returns list of vaults", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile1 := manager.config.Vaults["vault1"]
		profile2 := manager.config.Vaults["vault2"]
		manager.clients["vault1"] = createTestClient(t, &profile1)
		manager.clients["vault2"] = createTestClient(t, &profile2)

		vaults := manager.ListVaults()
		if len(vaults) != 2 {
			t.Errorf("Expected 2 vaults, got %d", len(vaults))
		}

		vaultMap := make(map[string]bool)
		for _, v := range vaults {
			vaultMap[v] = true
		}

		if !vaultMap["vault1"] || !vaultMap["vault2"] {
			t.Error("Expected vault1 and vault2 to be in the list")
		}
	})

	t.Run("returns empty list when no vaults", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		vaults := manager.ListVaults()
		if len(vaults) != 0 {
			t.Errorf("Expected 0 vaults, got %d", len(vaults))
		}
	})
}

func TestManager_GetActiveVault(t *testing.T) {
	t.Run("returns active vault name", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)
		manager.activeVault = "vault1"

		activeVault := manager.GetActiveVault()
		if activeVault != "vault1" {
			t.Errorf("Expected 'vault1', got '%s'", activeVault)
		}
	})

	t.Run("returns empty string when no active vault", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)
		manager.activeVault = ""

		activeVault := manager.GetActiveVault()
		if activeVault != "" {
			t.Errorf("Expected empty string, got '%s'", activeVault)
		}
	})
}

func TestManager_GetConnectionManager(t *testing.T) {
	t.Run("returns connection manager", func(t *testing.T) {
		manager, mockConnMgr, _, _ := createTestManager(t)

		connMgr := manager.GetConnectionManager()
		if connMgr != mockConnMgr {
			t.Error("Expected to get the connection manager")
		}
	})
}

func TestManager_GetSecretsManager(t *testing.T) {
	t.Run("returns secrets manager for active vault", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile := manager.config.Vaults["vault1"]
		client := createTestClient(t, &profile)
		manager.clients["vault1"] = client
		manager.activeVault = "vault1"

		secretsMgr, err := manager.GetSecretsManager()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if secretsMgr == nil {
			t.Error("Expected secrets manager to be returned")
		}
	})

	t.Run("returns error when no active client", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)
		manager.activeVault = "non-existent"

		_, err := manager.GetSecretsManager()
		if err == nil {
			t.Error("Expected error when no active client")
		}
	})
}

func TestManager_RefreshConnections(t *testing.T) {
	t.Run("calls RefreshAllConnections", func(t *testing.T) {
		manager, mockConnMgr, _, _ := createTestManager(t)

		manager.RefreshConnections()

		if mockConnMgr.RefreshAllConnectionsCalls != 1 {
			t.Errorf("Expected RefreshAllConnections to be called once, got %d calls", mockConnMgr.RefreshAllConnectionsCalls)
		}
	})
}

func TestManager_GetConnectionStatus(t *testing.T) {
	t.Run("returns connection status", func(t *testing.T) {
		manager, mockConnMgr, _, _ := createTestManager(t)

		expectedStatus := &ConnectionStatus{
			Connected: true,
			Address:   "http://vault1.local:8200",
		}
		mockConnMgr.ConnectionStatuses["vault1"] = expectedStatus

		status, err := manager.GetConnectionStatus("vault1")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if status != expectedStatus {
			t.Error("Expected to get the connection status")
		}

		if mockConnMgr.GetConnectionStatusCalls != 1 {
			t.Errorf("Expected GetConnectionStatus to be called once, got %d calls", mockConnMgr.GetConnectionStatusCalls)
		}
	})
}

func TestManager_GetHealthyConnections(t *testing.T) {
	t.Run("returns healthy connections", func(t *testing.T) {
		manager, mockConnMgr, _, _ := createTestManager(t)

		mockConnMgr.SetupConnection("vault1", true, false)
		mockConnMgr.SetupConnection("vault2", true, true)
		mockConnMgr.SetupConnection("vault3", false, false)

		healthy := manager.GetHealthyConnections()

		if len(healthy) != 1 {
			t.Errorf("Expected 1 healthy connection, got %d", len(healthy))
		}

		if len(healthy) > 0 && healthy[0] != "vault1" {
			t.Errorf("Expected vault1 to be healthy, got %s", healthy[0])
		}

		if mockConnMgr.GetHealthyConnectionsCalls != 1 {
			t.Errorf("Expected GetHealthyConnections to be called once, got %d calls", mockConnMgr.GetHealthyConnectionsCalls)
		}
	})
}

func TestManager_GetConnectedConnections(t *testing.T) {
	t.Run("returns connected connections", func(t *testing.T) {
		manager, mockConnMgr, _, _ := createTestManager(t)

		mockConnMgr.SetupConnection("vault1", true, false)
		mockConnMgr.SetupConnection("vault2", true, true)
		mockConnMgr.SetupConnection("vault3", false, false)

		connected := manager.GetConnectedConnections()

		if len(connected) != 2 {
			t.Errorf("Expected 2 connected connections, got %d", len(connected))
		}

		if mockConnMgr.GetConnectedConnectionsCalls != 1 {
			t.Errorf("Expected GetConnectedConnections to be called once, got %d calls", mockConnMgr.GetConnectedConnectionsCalls)
		}
	})
}

func TestManager_GetVaultProfiles(t *testing.T) {
	t.Run("returns all vault profiles", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profiles := manager.GetVaultProfiles()

		if len(profiles) != 2 {
			t.Errorf("Expected 2 profiles, got %d", len(profiles))
		}

		if _, exists := profiles["vault1"]; !exists {
			t.Error("Expected vault1 profile to exist")
		}

		if _, exists := profiles["vault2"]; !exists {
			t.Error("Expected vault2 profile to exist")
		}
	})
}

func TestManager_GetVaultProfile(t *testing.T) {
	t.Run("returns specific vault profile", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile, err := manager.GetVaultProfile("vault1")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if profile == nil {
			t.Fatal("Expected profile to be returned")
		}

		if profile.Address != "http://vault1.local:8200" {
			t.Errorf("Expected address 'http://vault1.local:8200', got '%s'", profile.Address)
		}
	})

	t.Run("returns error when profile not found", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		_, err := manager.GetVaultProfile("non-existent")
		if err == nil {
			t.Error("Expected error when profile not found")
		}
	})
}

func TestManager_GetVaultStatus(t *testing.T) {
	t.Run("returns status for all vaults", func(t *testing.T) {
		manager, mockConnMgr, _, _ := createTestManager(t)

		profile1 := manager.config.Vaults["vault1"]
		profile2 := manager.config.Vaults["vault2"]
		manager.clients["vault1"] = createTestClient(t, &profile1)
		manager.clients["vault2"] = createTestClient(t, &profile2)
		manager.activeVault = "vault1"

		mockConnMgr.SetupConnection("vault1", true, false)
		mockConnMgr.SetupConnection("vault2", false, false)

		statuses := manager.GetVaultStatus()

		if len(statuses) != 2 {
			t.Errorf("Expected 2 statuses, got %d", len(statuses))
		}

		vault1Status, exists := statuses["vault1"]
		if !exists {
			t.Fatal("Expected vault1 status to exist")
		}

		if vault1Status.Name != "vault1" {
			t.Errorf("Expected name 'vault1', got '%s'", vault1Status.Name)
		}

		if !vault1Status.Active {
			t.Error("Expected vault1 to be active")
		}

		if !vault1Status.Connected {
			t.Error("Expected vault1 to be connected")
		}

		vault2Status, exists := statuses["vault2"]
		if !exists {
			t.Fatal("Expected vault2 status to exist")
		}

		if vault2Status.Active {
			t.Error("Expected vault2 to not be active")
		}

		if vault2Status.Connected {
			t.Error("Expected vault2 to not be connected")
		}
	})

	t.Run("handles connection status errors gracefully", func(t *testing.T) {
		manager, mockConnMgr, _, _ := createTestManager(t)

		profile := manager.config.Vaults["vault1"]
		manager.clients["vault1"] = createTestClient(t, &profile)

		mockConnMgr.GetConnectionStatusFunc = func(name string) (*ConnectionStatus, error) {
			return nil, fmt.Errorf("connection error")
		}

		statuses := manager.GetVaultStatus()

		if len(statuses) != 1 {
			t.Errorf("Expected 1 status, got %d", len(statuses))
		}

		vault1Status, exists := statuses["vault1"]
		if !exists {
			t.Fatal("Expected vault1 status to exist")
		}

		if vault1Status.Connected {
			t.Error("Expected vault1 to not be connected")
		}

		if vault1Status.Error != "connection error" {
			t.Errorf("Expected error 'connection error', got '%s'", vault1Status.Error)
		}
	})
}

func TestManager_CreateClient(t *testing.T) {
	t.Run("creates client with basic config", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{
				Token: "test-token",
			},
		}

		client, err := manager.createClient(profile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if client == nil {
			t.Fatal("Expected client to be created")
		}

		if client.apiClient.Address() != "http://localhost:8200" {
			t.Errorf("Expected address 'http://localhost:8200', got '%s'", client.apiClient.Address())
		}
	})

	t.Run("creates client with namespace", func(t *testing.T) {
		manager, _, _, _ := createTestManager(t)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			Namespace:  "test-ns",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{
				Token: "test-token",
			},
		}

		client, err := manager.createClient(profile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Note: Can't easily test namespace setting as it's internal to api.Client
		if client == nil {
			t.Fatal("Expected client to be created")
		}
	})
}

func TestManager_InitializeProfileClient(t *testing.T) {
	t.Run("initializes profile client", func(t *testing.T) {
		manager, mockConnMgr, _, _ := createTestManager(t)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{
				Token: "test-token",
			},
		}

		err := manager.initializeProfileClient("test-vault", profile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if _, exists := manager.clients["test-vault"]; !exists {
			t.Error("Expected client to be added")
		}

		if mockConnMgr.AddConnectionCalls != 1 {
			t.Errorf("Expected AddConnection to be called once, got %d calls", mockConnMgr.AddConnectionCalls)
		}
	})
}

func TestManager_ReloadConfiguration(t *testing.T) {
	// Note: ReloadConfiguration depends on config.Load() which reads from the filesystem
	// and is not suitable for unit tests. This should be covered by integration tests.
	// However, we can test the helper functions it uses.
	t.Run("placeholder for integration test", func(t *testing.T) {
		t.Skip("ReloadConfiguration requires file system access and should be tested in integration tests")
	})
}

func TestProfileChanged(t *testing.T) {
	t.Run("detects address change", func(t *testing.T) {
		old := config.VaultProfile{
			Address:    "http://old.local:8200",
			AuthMethod: "token",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		new := config.VaultProfile{
			Address:    "http://new.local:8200",
			AuthMethod: "token",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}

		if !profileChanged(old, new) {
			t.Error("Expected profile to be detected as changed")
		}
	})

	t.Run("detects auth method change", func(t *testing.T) {
		old := config.VaultProfile{
			Address:    "http://vault.local:8200",
			AuthMethod: "token",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		new := config.VaultProfile{
			Address:    "http://vault.local:8200",
			AuthMethod: "userpass",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Username: "user", Password: "pass"},
		}

		if !profileChanged(old, new) {
			t.Error("Expected profile to be detected as changed")
		}
	})

	t.Run("detects namespace change", func(t *testing.T) {
		old := config.VaultProfile{
			Address:    "http://vault.local:8200",
			AuthMethod: "token",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		new := config.VaultProfile{
			Address:    "http://vault.local:8200",
			AuthMethod: "token",
			Namespace:  "ns2",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}

		if !profileChanged(old, new) {
			t.Error("Expected profile to be detected as changed")
		}
	})

	t.Run("detects auth config change", func(t *testing.T) {
		old := config.VaultProfile{
			Address:    "http://vault.local:8200",
			AuthMethod: "token",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		new := config.VaultProfile{
			Address:    "http://vault.local:8200",
			AuthMethod: "token",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Token: "token2"},
		}

		if !profileChanged(old, new) {
			t.Error("Expected profile to be detected as changed")
		}
	})

	t.Run("detects no change when profiles are identical", func(t *testing.T) {
		old := config.VaultProfile{
			Address:    "http://vault.local:8200",
			AuthMethod: "token",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		new := config.VaultProfile{
			Address:    "http://vault.local:8200",
			AuthMethod: "token",
			Namespace:  "ns1",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}

		if profileChanged(old, new) {
			t.Error("Expected profile to not be detected as changed")
		}
	})
}

func TestAuthConfigEqual(t *testing.T) {
	t.Run("detects equal auth configs", func(t *testing.T) {
		a := config.AuthConfig{Token: "token1"}
		b := config.AuthConfig{Token: "token1"}

		if !authConfigEqual(a, b) {
			t.Error("Expected auth configs to be equal")
		}
	})

	t.Run("detects different auth configs", func(t *testing.T) {
		a := config.AuthConfig{Token: "token1"}
		b := config.AuthConfig{Token: "token2"}

		if authConfigEqual(a, b) {
			t.Error("Expected auth configs to be different")
		}
	})
}
