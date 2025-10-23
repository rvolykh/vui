package vault

import (
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rvolykh/vui/internal/config"
	"github.com/sirupsen/logrus"
)

func TestNewConnectionManager(t *testing.T) {
	t.Run("creates connection manager with logger", func(t *testing.T) {
		logger := logrus.New()
		cm := NewConnectionManager(logger)

		if cm == nil {
			t.Fatal("Expected connection manager to be created")
		}

		if cm.logger != logger {
			t.Error("Expected logger to be set")
		}

		if cm.clients == nil {
			t.Error("Expected clients map to be initialized")
		}

		if cm.status == nil {
			t.Error("Expected status map to be initialized")
		}

		if len(cm.clients) != 0 {
			t.Errorf("Expected clients map to be empty, got %d entries", len(cm.clients))
		}

		if len(cm.status) != 0 {
			t.Errorf("Expected status map to be empty, got %d entries", len(cm.status))
		}
	})
}

func TestConnectionManager_AddConnection(t *testing.T) {
	t.Run("adds connection successfully", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "test-token"},
		}
		client := createTestClient(t, profile)

		cm.AddConnection("test-vault", client)

		if len(cm.clients) != 1 {
			t.Errorf("Expected 1 client, got %d", len(cm.clients))
		}

		if len(cm.status) != 1 {
			t.Errorf("Expected 1 status entry, got %d", len(cm.status))
		}

		retrievedClient, exists := cm.clients["test-vault"]
		if !exists {
			t.Fatal("Expected client to exist")
		}

		if retrievedClient != client {
			t.Error("Expected to retrieve the same client")
		}

		status, exists := cm.status["test-vault"]
		if !exists {
			t.Fatal("Expected status to exist")
		}

		if !status.Connecting {
			t.Error("Expected status to be connecting")
		}

		if status.Address != "http://localhost:8200" {
			t.Errorf("Expected address 'http://localhost:8200', got '%s'", status.Address)
		}
	})

	t.Run("overwrites existing connection", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile1 := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		client1 := createTestClient(t, profile1)

		profile2 := &config.VaultProfile{
			Address:    "http://localhost:8201",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token2"},
		}
		client2 := createTestClient(t, profile2)

		cm.AddConnection("test-vault", client1)
		cm.AddConnection("test-vault", client2)

		if len(cm.clients) != 1 {
			t.Errorf("Expected 1 client, got %d", len(cm.clients))
		}

		retrievedClient := cm.clients["test-vault"]
		if retrievedClient != client2 {
			t.Error("Expected second client to overwrite first")
		}

		status := cm.status["test-vault"]
		if status.Address != "http://localhost:8201" {
			t.Errorf("Expected address 'http://localhost:8201', got '%s'", status.Address)
		}
	})
}

func TestConnectionManager_RemoveConnection(t *testing.T) {
	t.Run("removes connection successfully", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "test-token"},
		}
		client := createTestClient(t, profile)

		cm.AddConnection("test-vault", client)
		cm.RemoveConnection("test-vault")

		if len(cm.clients) != 0 {
			t.Errorf("Expected 0 clients, got %d", len(cm.clients))
		}

		if len(cm.status) != 0 {
			t.Errorf("Expected 0 status entries, got %d", len(cm.status))
		}
	})

	t.Run("removing non-existent connection is safe", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		// Should not panic
		cm.RemoveConnection("non-existent")

		if len(cm.clients) != 0 {
			t.Errorf("Expected 0 clients, got %d", len(cm.clients))
		}
	})

	t.Run("removes only specified connection", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile1 := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		client1 := createTestClient(t, profile1)

		profile2 := &config.VaultProfile{
			Address:    "http://localhost:8201",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token2"},
		}
		client2 := createTestClient(t, profile2)

		cm.AddConnection("vault1", client1)
		cm.AddConnection("vault2", client2)
		cm.RemoveConnection("vault1")

		if len(cm.clients) != 1 {
			t.Errorf("Expected 1 client, got %d", len(cm.clients))
		}

		if _, exists := cm.clients["vault2"]; !exists {
			t.Error("Expected vault2 to still exist")
		}

		if _, exists := cm.clients["vault1"]; exists {
			t.Error("Expected vault1 to be removed")
		}
	})
}

func TestConnectionManager_GetConnection(t *testing.T) {
	t.Run("returns connection when it exists", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "test-token"},
		}
		client := createTestClient(t, profile)

		cm.AddConnection("test-vault", client)

		retrievedClient, err := cm.GetConnection("test-vault")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if retrievedClient != client {
			t.Error("Expected to retrieve the same client")
		}
	})

	t.Run("returns error when connection not found", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		_, err := cm.GetConnection("non-existent")
		if err == nil {
			t.Error("Expected error when connection not found")
		}

		expectedMsg := "connection 'non-existent' not found"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestConnectionManager_GetConnectionStatus(t *testing.T) {
	t.Run("returns status when connection exists", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "test-token"},
		}
		client := createTestClient(t, profile)

		cm.AddConnection("test-vault", client)

		status, err := cm.GetConnectionStatus("test-vault")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if status == nil {
			t.Fatal("Expected status to be returned")
		}

		if !status.Connecting {
			t.Error("Expected status to be connecting")
		}

		if status.Address != "http://localhost:8200" {
			t.Errorf("Expected address 'http://localhost:8200', got '%s'", status.Address)
		}
	})

	t.Run("returns copy of status to prevent race conditions", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "test-token"},
		}
		client := createTestClient(t, profile)

		cm.AddConnection("test-vault", client)

		status1, _ := cm.GetConnectionStatus("test-vault")
		status2, _ := cm.GetConnectionStatus("test-vault")

		// Modify status1
		status1.Connected = true
		status1.Error = "test error"

		// status2 should not be affected
		if status2.Connected {
			t.Error("Expected status2 to not be affected by changes to status1")
		}

		if status2.Error != "" {
			t.Error("Expected status2 error to be empty")
		}
	})

	t.Run("returns error when connection not found", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		_, err := cm.GetConnectionStatus("non-existent")
		if err == nil {
			t.Error("Expected error when connection not found")
		}

		expectedMsg := "connection 'non-existent' not found"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestConnectionManager_ListConnections(t *testing.T) {
	t.Run("returns empty list when no connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		connections := cm.ListConnections()
		if len(connections) != 0 {
			t.Errorf("Expected 0 connections, got %d", len(connections))
		}
	})

	t.Run("returns all connection names", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile1 := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		client1 := createTestClient(t, profile1)

		profile2 := &config.VaultProfile{
			Address:    "http://localhost:8201",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token2"},
		}
		client2 := createTestClient(t, profile2)

		profile3 := &config.VaultProfile{
			Address:    "http://localhost:8202",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token3"},
		}
		client3 := createTestClient(t, profile3)

		cm.AddConnection("vault1", client1)
		cm.AddConnection("vault2", client2)
		cm.AddConnection("vault3", client3)

		connections := cm.ListConnections()
		if len(connections) != 3 {
			t.Errorf("Expected 3 connections, got %d", len(connections))
		}

		connectionMap := make(map[string]bool)
		for _, name := range connections {
			connectionMap[name] = true
		}

		if !connectionMap["vault1"] || !connectionMap["vault2"] || !connectionMap["vault3"] {
			t.Error("Expected vault1, vault2, and vault3 to be in the list")
		}
	})
}

func TestConnectionManager_SetAllConnecting(t *testing.T) {
	t.Run("sets all connections to connecting state", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile1 := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token1"},
		}
		client1 := createTestClient(t, profile1)

		profile2 := &config.VaultProfile{
			Address:    "http://localhost:8201",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token2"},
		}
		client2 := createTestClient(t, profile2)

		cm.AddConnection("vault1", client1)
		cm.AddConnection("vault2", client2)

		// Set some statuses to non-connecting
		cm.status["vault1"].Connecting = false
		cm.status["vault1"].Connected = true
		cm.status["vault2"].Connecting = false
		cm.status["vault2"].Error = "some error"

		cm.SetAllConnecting()

		for name, status := range cm.status {
			if !status.Connecting {
				t.Errorf("Expected %s to be connecting", name)
			}
			if status.Error != "" {
				t.Errorf("Expected %s error to be cleared, got '%s'", name, status.Error)
			}
		}
	})

	t.Run("does nothing when no connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		// Should not panic
		cm.SetAllConnecting()

		if len(cm.status) != 0 {
			t.Errorf("Expected 0 status entries, got %d", len(cm.status))
		}
	})
}

func TestConnectionManager_GetHealthyConnections(t *testing.T) {
	t.Run("returns only healthy connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}

		// Add multiple connections with different states
		cm.AddConnection("healthy", createTestClient(t, profile))
		cm.AddConnection("sealed", createTestClient(t, profile))
		cm.AddConnection("not-initialized", createTestClient(t, profile))
		cm.AddConnection("disconnected", createTestClient(t, profile))

		// Set statuses
		cm.status["healthy"] = &ConnectionStatus{
			Connected:   true,
			Sealed:      false,
			Initialized: true,
		}
		cm.status["sealed"] = &ConnectionStatus{
			Connected:   true,
			Sealed:      true,
			Initialized: true,
		}
		cm.status["not-initialized"] = &ConnectionStatus{
			Connected:   true,
			Sealed:      false,
			Initialized: false,
		}
		cm.status["disconnected"] = &ConnectionStatus{
			Connected:   false,
			Sealed:      false,
			Initialized: true,
		}

		healthy := cm.GetHealthyConnections()

		if len(healthy) != 1 {
			t.Errorf("Expected 1 healthy connection, got %d", len(healthy))
		}

		if len(healthy) > 0 && healthy[0] != "healthy" {
			t.Errorf("Expected 'healthy' to be in the list, got '%s'", healthy[0])
		}
	})

	t.Run("returns empty list when no healthy connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}

		cm.AddConnection("sealed", createTestClient(t, profile))
		cm.status["sealed"] = &ConnectionStatus{
			Connected:   true,
			Sealed:      true,
			Initialized: true,
		}

		healthy := cm.GetHealthyConnections()

		if len(healthy) != 0 {
			t.Errorf("Expected 0 healthy connections, got %d", len(healthy))
		}
	})

	t.Run("returns empty list when no connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		healthy := cm.GetHealthyConnections()

		if len(healthy) != 0 {
			t.Errorf("Expected 0 healthy connections, got %d", len(healthy))
		}
	})
}

func TestConnectionManager_GetConnectedConnections(t *testing.T) {
	t.Run("returns all connected connections regardless of status", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}

		// Add multiple connections with different states
		cm.AddConnection("connected-healthy", createTestClient(t, profile))
		cm.AddConnection("connected-sealed", createTestClient(t, profile))
		cm.AddConnection("connected-not-init", createTestClient(t, profile))
		cm.AddConnection("disconnected", createTestClient(t, profile))

		// Set statuses
		cm.status["connected-healthy"] = &ConnectionStatus{
			Connected:   true,
			Sealed:      false,
			Initialized: true,
		}
		cm.status["connected-sealed"] = &ConnectionStatus{
			Connected:   true,
			Sealed:      true,
			Initialized: true,
		}
		cm.status["connected-not-init"] = &ConnectionStatus{
			Connected:   true,
			Sealed:      false,
			Initialized: false,
		}
		cm.status["disconnected"] = &ConnectionStatus{
			Connected:   false,
			Sealed:      false,
			Initialized: true,
		}

		connected := cm.GetConnectedConnections()

		if len(connected) != 3 {
			t.Errorf("Expected 3 connected connections, got %d", len(connected))
		}

		connMap := make(map[string]bool)
		for _, name := range connected {
			connMap[name] = true
		}

		if !connMap["connected-healthy"] {
			t.Error("Expected 'connected-healthy' to be in the list")
		}
		if !connMap["connected-sealed"] {
			t.Error("Expected 'connected-sealed' to be in the list")
		}
		if !connMap["connected-not-init"] {
			t.Error("Expected 'connected-not-init' to be in the list")
		}
		if connMap["disconnected"] {
			t.Error("Expected 'disconnected' to not be in the list")
		}
	})

	t.Run("returns empty list when no connected connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}

		cm.AddConnection("disconnected", createTestClient(t, profile))
		cm.status["disconnected"] = &ConnectionStatus{
			Connected: false,
		}

		connected := cm.GetConnectedConnections()

		if len(connected) != 0 {
			t.Errorf("Expected 0 connected connections, got %d", len(connected))
		}
	})

	t.Run("returns empty list when no connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		connected := cm.GetConnectedConnections()

		if len(connected) != 0 {
			t.Errorf("Expected 0 connected connections, got %d", len(connected))
		}
	})
}

func TestConnectionManager_TestConnectionAsync(t *testing.T) {
	t.Run("does nothing when connection not found", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		// Should not panic
		cm.TestConnectionAsync("non-existent")

		// Give it a moment to potentially start goroutine
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("starts async connection test for existing connection", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}
		client := createTestClient(t, profile)

		cm.AddConnection("test-vault", client)

		// Verify initial state
		status, _ := cm.GetConnectionStatus("test-vault")
		if !status.Connecting {
			t.Error("Expected initial status to be connecting")
		}

		// Call TestConnectionAsync - it will fail because vault is not running
		cm.TestConnectionAsync("test-vault")

		// Wait for the goroutine to complete (with timeout)
		connectionTimeout = 1 * time.Second
		maxWait := 2 * time.Second
		checkInterval := 100 * time.Millisecond
		elapsed := time.Duration(0)

		for elapsed < maxWait {
			time.Sleep(checkInterval)
			elapsed += checkInterval

			status, _ = cm.GetConnectionStatus("test-vault")
			if !status.Connecting {
				// Status has been updated
				break
			}
		}

		// Verify the status was updated (should be disconnected with error)
		status, _ = cm.GetConnectionStatus("test-vault")
		if status.Connecting {
			t.Error("Expected status to not be connecting after test")
		}
		if status.Connected {
			t.Error("Expected status to be disconnected (vault not running)")
		}
		if status.Error == "" {
			t.Error("Expected error to be set (vault not running)")
		}
	})
}

func TestConnectionManager_RefreshConnectionStatus(t *testing.T) {
	t.Run("returns error when connection not found", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		err := cm.RefreshConnectionStatus("non-existent")
		if err == nil {
			t.Error("Expected error when connection not found")
		}

		expectedMsg := "connection 'non-existent' not found"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("returns error when vault is not reachable", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}
		client := createTestClient(t, profile)

		cm.AddConnection("test-vault", client)

		err := cm.RefreshConnectionStatus("test-vault")
		if err == nil {
			t.Error("Expected error when vault is not reachable")
		}

		// Verify status was updated with error
		status, _ := cm.GetConnectionStatus("test-vault")
		if status.Connected {
			t.Error("Expected status to be disconnected")
		}
		if status.Error == "" {
			t.Error("Expected error to be set")
		}
	})
}

func TestConnectionManager_RefreshAllConnections(t *testing.T) {
	t.Run("refreshes all connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}

		cm.AddConnection("vault1", createTestClient(t, profile))
		cm.AddConnection("vault2", createTestClient(t, profile))
		cm.AddConnection("vault3", createTestClient(t, profile))

		// Set some initial states
		cm.status["vault1"].Connected = true
		cm.status["vault2"].Connected = true
		cm.status["vault3"].Connected = false
		cm.status["vault3"].Connecting = false

		// Call RefreshAllConnections
		cm.RefreshAllConnections()

		// Wait for goroutines to complete (with timeout)
		connectionTimeout = 1 * time.Second
		maxWait := 2 * time.Second
		checkInterval := 100 * time.Millisecond
		elapsed := time.Duration(0)

		for elapsed < maxWait {
			time.Sleep(checkInterval)
			elapsed += checkInterval

			status1, _ := cm.GetConnectionStatus("vault1")
			status2, _ := cm.GetConnectionStatus("vault2")

			// Check if both connections have finished their async tests
			if !status1.Connecting && !status2.Connecting {
				break
			}
		}

		// vault1 and vault2 should be refreshed (they were connected)
		// vault3 should not be refreshed (it was disconnected and not connecting)
		// Since vault is not running, they should all end up disconnected
		status1, _ := cm.GetConnectionStatus("vault1")
		status2, _ := cm.GetConnectionStatus("vault2")

		if status1.Connecting {
			t.Error("Expected vault1 to not be connecting after refresh")
		}
		if status2.Connecting {
			t.Error("Expected vault2 to not be connecting after refresh")
		}

		// Verify errors are set
		if status1.Error == "" {
			t.Error("Expected vault1 to have an error (vault not running)")
		}
		if status2.Error == "" {
			t.Error("Expected vault2 to have an error (vault not running)")
		}
	})

	t.Run("does nothing when no connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		// Should not panic
		cm.RefreshAllConnections()

		time.Sleep(50 * time.Millisecond)
	})

	t.Run("skips failed connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}

		cm.AddConnection("failed-vault", createTestClient(t, profile))

		// Set to failed state (not connected, not connecting)
		cm.status["failed-vault"].Connected = false
		cm.status["failed-vault"].Connecting = false
		cm.status["failed-vault"].Error = "previous error"

		initialError := cm.status["failed-vault"].Error

		cm.RefreshAllConnections()

		time.Sleep(50 * time.Millisecond)

		// Error should remain the same (connection was not refreshed)
		status, _ := cm.GetConnectionStatus("failed-vault")
		if status.Error != initialError {
			t.Error("Expected error to remain unchanged for failed connection")
		}
	})
}

func TestConnectionStatus(t *testing.T) {
	t.Run("can create connection status with all fields", func(t *testing.T) {
		now := time.Now()
		status := &ConnectionStatus{
			Connecting:  true,
			Connected:   false,
			Address:     "http://localhost:8200",
			Sealed:      true,
			Initialized: false,
			Version:     "1.0.0",
			ClusterID:   "cluster-123",
			LastCheck:   now,
			Error:       "test error",
		}

		if !status.Connecting {
			t.Error("Expected Connecting to be true")
		}
		if status.Connected {
			t.Error("Expected Connected to be false")
		}
		if status.Address != "http://localhost:8200" {
			t.Errorf("Expected Address 'http://localhost:8200', got '%s'", status.Address)
		}
		if !status.Sealed {
			t.Error("Expected Sealed to be true")
		}
		if status.Initialized {
			t.Error("Expected Initialized to be false")
		}
		if status.Version != "1.0.0" {
			t.Errorf("Expected Version '1.0.0', got '%s'", status.Version)
		}
		if status.ClusterID != "cluster-123" {
			t.Errorf("Expected ClusterID 'cluster-123', got '%s'", status.ClusterID)
		}
		if !status.LastCheck.Equal(now) {
			t.Error("Expected LastCheck to match")
		}
		if status.Error != "test error" {
			t.Errorf("Expected Error 'test error', got '%s'", status.Error)
		}
	})
}

func TestConnectionManager_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent reads and writes safely", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		profile := &config.VaultProfile{
			Address:    "http://localhost:8200",
			AuthMethod: "token",
			AuthConfig: config.AuthConfig{Token: "token"},
		}

		// Add initial connection
		cm.AddConnection("test-vault", createTestClient(t, profile))

		// Spawn multiple goroutines that access the connection manager
		done := make(chan bool, 10)

		// Readers
		for i := 0; i < 5; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					cm.GetConnection("test-vault")
					cm.GetConnectionStatus("test-vault")
					cm.ListConnections()
					cm.GetHealthyConnections()
					cm.GetConnectedConnections()
					time.Sleep(time.Millisecond)
				}
				done <- true
			}()
		}

		// Writers
		for i := 0; i < 5; i++ {
			go func(id int) {
				for j := 0; j < 10; j++ {
					name := "vault-" + string(rune('0'+id))
					cm.AddConnection(name, createTestClient(t, profile))
					cm.SetAllConnecting()
					cm.RemoveConnection(name)
					time.Sleep(time.Millisecond)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Helper function for creating test API client (used across tests)
func createTestAPIClient(address string) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = address
	return api.NewClient(config)
}
