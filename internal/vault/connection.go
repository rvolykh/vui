package vault

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ConnectionStatus represents the status of a vault connection
type ConnectionStatus struct {
	Connecting  bool      `json:"connecting"`
	Connected   bool      `json:"connected"`
	Address     string    `json:"address"`
	Sealed      bool      `json:"sealed"`
	Initialized bool      `json:"initialized"`
	Version     string    `json:"version"`
	ClusterID   string    `json:"cluster_id"`
	LastCheck   time.Time `json:"last_check"`
	Error       string    `json:"error,omitempty"`
}

// ConnectionManager manages vault connections and their status
type ConnectionManager struct {
	clients map[string]*Client
	status  map[string]*ConnectionStatus
	mutex   sync.RWMutex
	logger  *logrus.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(logger *logrus.Logger) *ConnectionManager {
	return &ConnectionManager{
		clients: make(map[string]*Client),
		status:  make(map[string]*ConnectionStatus),
		logger:  logger,
	}
}

// AddConnection adds a new vault connection and sets its status to connecting
func (cm *ConnectionManager) AddConnection(name string, client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.clients[name] = client
	cm.status[name] = &ConnectionStatus{
		Connecting: true,
		Address:    client.apiClient.Address(),
		LastCheck:  time.Now(),
	}
}

// TestConnectionAsync tests a vault connection asynchronously
func (cm *ConnectionManager) TestConnectionAsync(name string) {
	cm.mutex.RLock()
	client, exists := cm.clients[name]
	cm.mutex.RUnlock()

	if !exists {
		return
	}

	go func() {
		status, err := cm.testConnection(client)
		if err != nil {
			cm.logger.Warnf("Failed to connect to vault '%s': %v", name, err)
			cm.mutex.Lock()
			if existingStatus, ok := cm.status[name]; ok {
				existingStatus.Connecting = false
				existingStatus.Connected = false
				existingStatus.Error = err.Error()
				existingStatus.LastCheck = time.Now()
			}
			cm.mutex.Unlock()
			return
		}

		cm.mutex.Lock()
		cm.status[name] = status
		cm.status[name].Connecting = false
		cm.mutex.Unlock()
		cm.logger.Infof("Updated vault connection status: %s (connected: %v)", name, status.Connected)
	}()
}

// RemoveConnection removes a vault connection
func (cm *ConnectionManager) RemoveConnection(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.clients, name)
	delete(cm.status, name)
	cm.logger.Infof("Removed vault connection: %s", name)
}

// GetConnection returns a connection by name
func (cm *ConnectionManager) GetConnection(name string) (*Client, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}

	return client, nil
}

// GetConnectionStatus returns the status of a connection
func (cm *ConnectionManager) GetConnectionStatus(name string) (*ConnectionStatus, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	status, exists := cm.status[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}

	// Return a copy to prevent race conditions
	statusCopy := *status
	return &statusCopy, nil
}

// ListConnections returns all connection names
func (cm *ConnectionManager) ListConnections() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		connections = append(connections, name)
	}

	return connections
}

// RefreshConnectionStatus refreshes the status of a specific connection
func (cm *ConnectionManager) RefreshConnectionStatus(name string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	client, exists := cm.clients[name]
	if !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}

	status, err := cm.testConnection(client)
	if err != nil {
		// Update status with error
		if existingStatus, ok := cm.status[name]; ok {
			existingStatus.Connected = false
			existingStatus.Error = err.Error()
			existingStatus.LastCheck = time.Now()
		}
		return err
	}

	cm.status[name] = status
	return nil
}

// RefreshAllConnections refreshes the status of all connections
func (cm *ConnectionManager) RefreshAllConnections() {
	cm.mutex.RLock()
	names := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		names = append(names, name)
	}
	cm.mutex.RUnlock()

	for _, name := range names {
		cm.mutex.RLock()
		status, ok := cm.status[name]
		cm.mutex.RUnlock()

		if ok && !status.Connected && !status.Connecting {
			// Do not refresh connections that have failed, unless manually triggered
			continue
		}
		cm.TestConnectionAsync(name)
	}
}

// SetAllConnecting sets all connections to "Connecting" state
func (cm *ConnectionManager) SetAllConnecting() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for name, status := range cm.status {
		status.Connecting = true
		status.Error = ""
		cm.logger.Debugf("Set connection '%s' to connecting state", name)
	}
}

// testConnection tests a vault connection and returns its status
func (cm *ConnectionManager) testConnection(client *Client) (*ConnectionStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get vault status
	status, err := client.apiClient.Sys().SealStatusWithContext(ctx)
	if err != nil {
		return &ConnectionStatus{
			Connected: false,
			Address:   client.apiClient.Address(),
			Error:     err.Error(),
			LastCheck: time.Now(),
		}, err
	}

	// Get vault health
	health, err := client.apiClient.Sys().HealthWithContext(ctx)
	if err != nil {
		return &ConnectionStatus{
			Connected: false,
			Address:   client.apiClient.Address(),
			Error:     err.Error(),
			LastCheck: time.Now(),
		}, err
	}

	return &ConnectionStatus{
		Connected:   true,
		Address:     client.apiClient.Address(),
		Sealed:      status.Sealed,
		Initialized: health.Initialized,
		Version:     status.Version,
		ClusterID:   status.ClusterID,
		LastCheck:   time.Now(),
	}, nil
}

// GetHealthyConnections returns connections that are healthy and connected
func (cm *ConnectionManager) GetHealthyConnections() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var healthy []string
	for name, status := range cm.status {
		if status.Connected && !status.Sealed && status.Initialized {
			healthy = append(healthy, name)
		}
	}

	return healthy
}

// GetConnectedConnections returns all connections that are connected (regardless of sealed/initialized status)
func (cm *ConnectionManager) GetConnectedConnections() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var connected []string
	for name, status := range cm.status {
		if status.Connected {
			connected = append(connected, name)
		}
	}

	return connected
}
