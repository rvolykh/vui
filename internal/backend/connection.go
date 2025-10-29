package backend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rvolykh/vui/internal/engines"
	"github.com/rvolykh/vui/internal/models"
	"github.com/sirupsen/logrus"
)

var (
	connectionTimeout = 10 * time.Second
)

type ConnectionManager struct {
	clients map[string]engines.SecretEngine
	status  map[string]*models.ConnectionStatus
	mutex   sync.RWMutex
	logger  *logrus.Logger
}

func NewConnectionManager(logger *logrus.Logger) *ConnectionManager {
	return &ConnectionManager{
		clients: make(map[string]engines.SecretEngine),
		status:  make(map[string]*models.ConnectionStatus),
		logger:  logger,
	}
}

func (cm *ConnectionManager) AddConnection(name string, client engines.SecretEngine) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.clients[name] = client
	cm.status[name] = &models.ConnectionStatus{
		Status:    models.StatusConnecting,
		Address:   client.GetAddress(),
		LastCheck: time.Now(),
	}
}

func (cm *ConnectionManager) TestConnectionAsync(name string) {
	cm.mutex.RLock()
	client, exists := cm.clients[name]
	cm.mutex.RUnlock()

	if !exists {
		return
	}

	go func() {
		status, err := testConnection(client)
		if err != nil {
			cm.logger.Warnf("Failed to connect to '%s': %v", name, err)
			cm.mutex.Lock()
			if existingStatus, ok := cm.status[name]; ok {
				existingStatus.Status = models.StatusDisconnected
				existingStatus.Error = err.Error()
				existingStatus.LastCheck = time.Now()
			}
			cm.mutex.Unlock()
			return
		}

		cm.mutex.Lock()
		cm.status[name] = &status
		cm.mutex.Unlock()
		cm.logger.Infof("Updated connection status: %s (%v)", name, status.Status)
	}()
}

func (cm *ConnectionManager) RemoveConnection(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.clients, name)
	delete(cm.status, name)
	cm.logger.Infof("Removed connection: %s", name)
}

func (cm *ConnectionManager) GetConnection(name string) (engines.SecretEngine, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[name]
	if !exists {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}

	return client, nil
}

func (cm *ConnectionManager) GetConnectionStatus(name string) (*models.ConnectionStatus, error) {
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

func (cm *ConnectionManager) ListConnections() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		connections = append(connections, name)
	}

	return connections
}

func (cm *ConnectionManager) RefreshConnectionStatus(name string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	client, exists := cm.clients[name]
	if !exists {
		return fmt.Errorf("connection '%s' not found", name)
	}

	status, err := testConnection(client)
	if err != nil {
		// Update status with error
		if existingStatus, ok := cm.status[name]; ok {
			existingStatus.Status = models.StatusDisconnected
			existingStatus.Error = err.Error()
			existingStatus.LastCheck = time.Now()
		}
		return err
	}

	cm.status[name] = &status
	return nil
}

func (cm *ConnectionManager) ResetConnections() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for name, status := range cm.status {
		status.Status = models.StatusConnecting
		status.Error = ""
		cm.logger.Debugf("Set connection '%s' to connecting state", name)
	}
}

func testConnection(client engines.SecretEngine) (models.ConnectionStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	status, err := client.GetStatus(ctx)
	if err != nil {
		return models.ConnectionStatus{
			Status:    models.StatusDisconnected,
			Address:   client.GetAddress(),
			Error:     err.Error(),
			LastCheck: time.Now(),
		}, err
	}

	return status, nil
}
