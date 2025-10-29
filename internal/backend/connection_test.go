package backend

import (
	"errors"
	"testing"
	"time"

	"github.com/rvolykh/vui/internal/engines/fake"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnectionManager(t *testing.T) {
	t.Run("creates connection manager with logger", func(t *testing.T) {
		logger := logrus.New()

		cm := NewConnectionManager(logger)

		assert.NotNil(t, cm, "Expected connection manager to be created")
		assert.Equal(t, logger, cm.logger, "Expected logger to be set")
		assert.NotNil(t, cm.clients, "Expected clients map to be initialized")
		assert.NotNil(t, cm.status, "Expected status map to be initialized")
		assert.Empty(t, cm.clients, "Expected clients map to be empty")
		assert.Empty(t, cm.status, "Expected status map to be empty")
	})
}

func TestConnectionManager_AddConnection(t *testing.T) {
	t.Run("adds connection successfully", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)
		client := fake.NewFakeClient()
		client.RespGetAddress = "http://localhost:8200"

		cm.AddConnection("test-vault", client)
		assert.Lenf(t, cm.clients, 1, "Expected 1 client, got %d", len(cm.clients))

		retrievedClient, exists := cm.clients["test-vault"]
		require.True(t, exists, "Expected client to exist")

		assert.Equal(t, client, retrievedClient, "Expected to retrieve the same client")

		status, exists := cm.status["test-vault"]
		require.True(t, exists, "Expected status to exist")

		assert.Equal(t, models.StatusConnecting, status.Status, "Expected status to be connecting")

		assert.Equal(t, "http://localhost:8200", status.Address, "Expected address 'http://localhost:8200'")
	})

	t.Run("overwrites existing connection", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client1 := fake.NewFakeClient()
		client1.RespGetAddress = "http://localhost:8200"

		client2 := fake.NewFakeClient()
		client2.RespGetAddress = "http://localhost:8201"

		cm.AddConnection("test-vault", client1)
		cm.AddConnection("test-vault", client2)

		assert.Lenf(t, cm.clients, 1, "Expected 1 client, got %d", len(cm.clients))

		retrievedClient := cm.clients["test-vault"]
		assert.Equal(t, client2, retrievedClient, "Expected second client to overwrite first")

		status := cm.status["test-vault"]
		assert.Equal(t, "http://localhost:8201", status.Address, "Expected address 'http://localhost:8201'")
	})
}

func TestConnectionManager_RemoveConnection(t *testing.T) {
	t.Run("removes connection successfully", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client := fake.NewFakeClient()
		client.RespGetAddress = "http://localhost:8200"

		cm.AddConnection("test-vault", client)
		cm.RemoveConnection("test-vault")

		assert.Lenf(t, cm.clients, 0, "Expected 0 clients, got %d", len(cm.clients))
		assert.Lenf(t, cm.status, 0, "Expected 0 status entries, got %d", len(cm.status))
	})

	t.Run("removing non-existent connection is safe", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		cm.RemoveConnection("non-existent")

		assert.Lenf(t, cm.clients, 0, "Expected 0 clients, got %d", len(cm.clients))
	})

	t.Run("removes only specified connection", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client1 := fake.NewFakeClient()
		client1.RespGetAddress = "http://localhost:8200"

		client2 := fake.NewFakeClient()
		client2.RespGetAddress = "http://localhost:8201"

		cm.AddConnection("vault1", client1)
		cm.AddConnection("vault2", client2)
		cm.RemoveConnection("vault1")

		assert.Lenf(t, cm.clients, 1, "Expected 1 client, got %d", len(cm.clients))

		_, exists := cm.clients["vault2"]
		assert.True(t, exists, "Expected vault2 to still exist")

		_, exists = cm.clients["vault1"]
		assert.False(t, exists, "Expected vault1 to be removed")
	})
}

func TestConnectionManager_GetConnection(t *testing.T) {
	t.Run("returns connection when it exists", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client := fake.NewFakeClient()
		client.RespGetAddress = "http://localhost:8200"

		cm.AddConnection("test-vault", client)

		retrievedClient, err := cm.GetConnection("test-vault")
		require.NoError(t, err, "Expected no error")

		assert.Equal(t, client, retrievedClient, "Expected to retrieve the same client")
	})

	t.Run("returns error when connection not found", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		_, err := cm.GetConnection("non-existent")
		require.Error(t, err, "Expected error when connection not found")

		assert.Equal(t, "connection 'non-existent' not found", err.Error(), "Expected error message 'connection 'non-existent' not found'")
	})
}

func TestConnectionManager_GetConnectionStatus(t *testing.T) {
	t.Run("returns status when connection exists", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client := fake.NewFakeClient()
		client.RespGetAddress = "http://localhost:8200"

		cm.AddConnection("test-vault", client)

		status, err := cm.GetConnectionStatus("test-vault")
		require.NoError(t, err, "Expected no error")

		assert.Equal(t, models.StatusConnecting, status.Status, "Expected status to be connecting")
		assert.Equal(t, "http://localhost:8200", status.Address, "Expected address 'http://localhost:8200'")
	})

	t.Run("returns copy of status to prevent race conditions", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client := fake.NewFakeClient()
		client.RespGetAddress = "http://localhost:8200"

		cm.AddConnection("test-vault", client)

		status1, _ := cm.GetConnectionStatus("test-vault")
		status2, _ := cm.GetConnectionStatus("test-vault")

		// Modify status1
		status1.Status = models.StatusConnected
		status1.Error = "test error"

		// status2 should not be affected
		assert.Equal(t, models.StatusConnecting, status2.Status, "Expected status2 to be connecting")
		assert.Empty(t, status2.Error, "Expected status2 error to be empty")
	})

	t.Run("returns error when connection not found", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		_, err := cm.GetConnectionStatus("non-existent")
		require.Error(t, err, "Expected error when connection not found")

		assert.Equal(t, "connection 'non-existent' not found", err.Error(), "Expected error message 'connection 'non-existent' not found'")
	})
}

func TestConnectionManager_ListConnections(t *testing.T) {
	t.Run("returns empty list when no connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		connections := cm.ListConnections()
		assert.Lenf(t, connections, 0, "Expected 0 connections, got %d", len(connections))
	})

	t.Run("returns all connection names", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client1 := fake.NewFakeClient()
		client1.RespGetAddress = "http://localhost:8200"

		client2 := fake.NewFakeClient()
		client2.RespGetAddress = "http://localhost:8201"

		client3 := fake.NewFakeClient()
		client3.RespGetAddress = "http://localhost:8202"

		cm.AddConnection("vault1", client1)
		cm.AddConnection("vault2", client2)
		cm.AddConnection("vault3", client3)

		connections := cm.ListConnections()
		assert.Lenf(t, connections, 3, "Expected 3 connections, got %d", len(connections))

		connectionMap := make(map[string]bool)
		for _, name := range connections {
			connectionMap[name] = true
		}

		assert.True(t, connectionMap["vault1"], "Expected vault1 to be in the list")
		assert.True(t, connectionMap["vault2"], "Expected vault2 to be in the list")
		assert.True(t, connectionMap["vault3"], "Expected vault3 to be in the list")
	})
}

func TestConnectionManager_ResetConnections(t *testing.T) {
	t.Run("sets all connections to connecting state", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client1 := fake.NewFakeClient()
		client1.RespGetAddress = "http://localhost:8200"

		client2 := fake.NewFakeClient()
		client2.RespGetAddress = "http://localhost:8201"

		cm.AddConnection("vault1", client1)
		cm.AddConnection("vault2", client2)

		// Set some statuses to non-connecting
		cm.status["vault1"].Status = models.StatusConnected
		cm.status["vault2"].Status = models.StatusConnecting
		cm.status["vault2"].Error = "some error"

		cm.ResetConnections()

		assert.Equal(t, models.StatusConnecting, cm.status["vault1"].Status, "Expected vault1 to be connecting")
		assert.Equal(t, models.StatusConnecting, cm.status["vault2"].Status, "Expected vault2 to be connecting")
		assert.Empty(t, cm.status["vault2"].Error, "Expected vault2 error to be empty")
	})

	t.Run("does nothing when no connections", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		// Should not panic
		cm.ResetConnections()

		assert.Lenf(t, cm.status, 0, "Expected 0 status entries, got %d", len(cm.status))
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

		client := fake.NewFakeClient()
		client.RespGetAddress = "http://localhost:8200"
		client.RespGetStatus = models.ConnectionStatus{
			Status: models.StatusDisconnected,
			Error:  "test error",
		}

		cm.AddConnection("test-vault", client)

		// Verify initial state
		status, _ := cm.GetConnectionStatus("test-vault")
		assert.Equal(t, models.StatusConnecting, status.Status, "Expected initial status to be connecting")

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
			if status.Status != models.StatusConnecting {
				// Status has been updated
				break
			}
		}

		// Verify the status was updated (should be disconnected with error)
		status, _ = cm.GetConnectionStatus("test-vault")
		assert.Equal(t, models.StatusDisconnected, status.Status, "Expected status to be disconnected")
		assert.NotEmpty(t, status.Error, "Expected error to be set")
	})
}

func TestConnectionManager_RefreshConnectionStatus(t *testing.T) {
	t.Run("returns error when connection not found", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		err := cm.RefreshConnectionStatus("non-existent")
		require.Error(t, err, "Expected error when connection not found")

		assert.Equal(t, "connection 'non-existent' not found", err.Error(), "Expected error message 'connection 'non-existent' not found'")
	})

	t.Run("returns error when vault is not reachable", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client := fake.NewFakeClient()
		client.RespGetAddress = "http://localhost:8200"
		client.RespErr = errors.New("test error")

		cm.AddConnection("test-vault", client)

		err := cm.RefreshConnectionStatus("test-vault")
		require.Error(t, err, "Expected error when vault is not reachable")

		// Verify status was updated with error
		status, _ := cm.GetConnectionStatus("test-vault")
		assert.Equal(t, models.StatusDisconnected, status.Status, "Expected status to be disconnected")
		assert.NotEmpty(t, status.Error, "Expected error to be set")
	})
}

func TestConnectionManager_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent reads and writes safely", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)
		cm := NewConnectionManager(logger)

		client := fake.NewFakeClient()
		client.RespGetAddress = "http://localhost:8200"

		// Add initial connection
		cm.AddConnection("test-vault", client)

		// Spawn multiple goroutines that access the connection manager
		done := make(chan bool, 10)

		// Readers
		for range 5 {
			go func() {
				for range 10 {
					cm.GetConnection("test-vault")
					cm.GetConnectionStatus("test-vault")
					cm.ListConnections()
					time.Sleep(time.Millisecond)
				}
				done <- true
			}()
		}

		// Writers
		for i := range 5 {
			go func(id int) {
				for range 10 {
					name := "vault-" + string(rune('0'+id))
					cm.AddConnection(name, fake.NewFakeClient())
					cm.RemoveConnection(name)
					time.Sleep(time.Millisecond)
				}
			}(i)
		}
	})
}
