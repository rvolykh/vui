package models

import "time"

type SecretNode struct {
	Name     string          `json:"name"`
	Path     string          `json:"path"`
	IsSecret bool            `json:"is_secret"`
	Children []*SecretNode   `json:"children,omitempty"`
	Data     map[string]any  `json:"data,omitempty"`
	Metadata *SecretMetadata `json:"metadata,omitempty"`
}

type SecretMetadata struct {
	CreatedTime  time.Time `json:"created_time"`
	Version      int       `json:"version"`
	Destroyed    bool      `json:"destroyed"`
	DeletionTime time.Time `json:"deletion_time,omitempty"`
}
