package models

import "time"

type ConnectionStatus struct {
	Status    Status    `json:"status"`
	Address   string    `json:"address"`
	Version   string    `json:"version"`
	ClusterID string    `json:"cluster_id"`
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
}
