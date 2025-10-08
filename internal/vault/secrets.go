package vault

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SecretNode represents a node in the secret tree
type SecretNode struct {
	Name     string                 `json:"name"`
	Path     string                 `json:"path"`
	IsSecret bool                   `json:"is_secret"`
	Children []*SecretNode          `json:"children,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Metadata *SecretMetadata        `json:"metadata,omitempty"`
}

// SecretMetadata contains metadata about a secret
type SecretMetadata struct {
	CreatedTime  time.Time `json:"created_time"`
	Version      int       `json:"version"`
	Destroyed    bool      `json:"destroyed"`
	DeletionTime time.Time `json:"deletion_time,omitempty"`
}

// SecretsManager manages secret operations
type SecretsManager struct {
	client *Client
	logger *logrus.Logger
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager(client *Client, logger *logrus.Logger) *SecretsManager {
	return &SecretsManager{
		client: client,
		logger: logger,
	}
}

// ListSecrets lists all secrets at a given path
func (sm *SecretsManager) ListSecrets(path string) ([]*SecretNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Normalize path
	path = strings.Trim(path, "/")
	if path == "" {
		path = ""
	}

	// List secrets at the path
	secret, err := sm.client.apiClient.Logical().ListWithContext(ctx, "secret/metadata/"+path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets at path '%s': %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return []*SecretNode{}, nil
	}

	// Extract keys from the response
	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []*SecretNode{}, nil
	}

	var nodes []*SecretNode
	for _, key := range keys {
		keyStr, ok := key.(string)
		if !ok {
			continue
		}

		// Remove trailing slash for directories
		keyStr = strings.TrimSuffix(keyStr, "/")

		node := &SecretNode{
			Name:     keyStr,
			Path:     filepath.Join(path, keyStr),
			IsSecret: !strings.HasSuffix(key.(string), "/"),
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetSecret retrieves a secret and its metadata
func (sm *SecretsManager) GetSecret(path string) (*SecretNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get secret data
	secret, err := sm.client.apiClient.Logical().ReadWithContext(ctx, "secret/data/"+path)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret at path '%s': %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path '%s'", path)
	}

	// Get secret metadata
	metadata, err := sm.client.apiClient.Logical().ReadWithContext(ctx, "secret/metadata/"+path)
	if err != nil {
		sm.logger.Warnf("Failed to get metadata for secret '%s': %v", path, err)
	}

	node := &SecretNode{
		Name:     filepath.Base(path),
		Path:     path,
		IsSecret: true,
	}

	// Extract data from KV v2 response
	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		node.Data = data
	} else {
		node.Data = secret.Data
	}

	// Extract metadata if available
	if metadata != nil && metadata.Data != nil {
		if createdTime, ok := metadata.Data["created_time"].(string); ok {
			if t, err := time.Parse(time.RFC3339, createdTime); err == nil {
				node.Metadata = &SecretMetadata{
					CreatedTime: t,
				}
			}
		}
		if version, ok := metadata.Data["current_version"].(float64); ok {
			if node.Metadata == nil {
				node.Metadata = &SecretMetadata{}
			}
			node.Metadata.Version = int(version)
		}
	}

	return node, nil
}

// CreateSecret creates a new secret
func (sm *SecretsManager) CreateSecret(path string, data map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// For KV v2, we need to wrap the data
	secretData := map[string]interface{}{
		"data": data,
	}

	_, err := sm.client.apiClient.Logical().WriteWithContext(ctx, "secret/data/"+path, secretData)
	if err != nil {
		return fmt.Errorf("failed to create secret at path '%s': %w", path, err)
	}

	sm.logger.Infof("Created secret at path: %s", path)
	return nil
}

// UpdateSecret updates an existing secret
func (sm *SecretsManager) UpdateSecret(path string, data map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// For KV v2, we need to wrap the data
	secretData := map[string]interface{}{
		"data": data,
	}

	_, err := sm.client.apiClient.Logical().WriteWithContext(ctx, "secret/data/"+path, secretData)
	if err != nil {
		return fmt.Errorf("failed to update secret at path '%s': %w", path, err)
	}

	sm.logger.Infof("Updated secret at path: %s", path)
	return nil
}

// DeleteSecret deletes a secret
func (sm *SecretsManager) DeleteSecret(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := sm.client.apiClient.Logical().DeleteWithContext(ctx, "secret/metadata/"+path)
	if err != nil {
		return fmt.Errorf("failed to delete secret at path '%s': %w", path, err)
	}

	sm.logger.Infof("Deleted secret at path: %s", path)
	return nil
}

// BuildTree builds a complete tree structure for a given path
func (sm *SecretsManager) BuildTree(rootPath string, maxDepth int) (*SecretNode, error) {
	return sm.buildTreeRecursive(rootPath, "", maxDepth, 0)
}

// buildTreeRecursive recursively builds the secret tree
func (sm *SecretsManager) buildTreeRecursive(rootPath, currentPath string, maxDepth, currentDepth int) (*SecretNode, error) {
	if currentDepth >= maxDepth {
		return nil, nil
	}

	// List secrets at current path
	secrets, err := sm.ListSecrets(currentPath)
	if err != nil {
		return nil, err
	}

	// Create root node
	node := &SecretNode{
		Name:     filepath.Base(currentPath),
		Path:     currentPath,
		IsSecret: false,
		Children: []*SecretNode{},
	}

	// If this is the root, use the provided name
	if currentPath == "" {
		node.Name = rootPath
		if rootPath == "" {
			node.Name = "secrets"
		}
	}

	// Process each secret/directory
	for _, secret := range secrets {
		if secret.IsSecret {
			// This is a secret, get its data
			secretNode, err := sm.GetSecret(secret.Path)
			if err != nil {
				sm.logger.Warnf("Failed to get secret '%s': %v", secret.Path, err)
				// Add node without data if we can't retrieve it
				node.Children = append(node.Children, secret)
				continue
			}
			node.Children = append(node.Children, secretNode)
		} else {
			// This is a directory, recurse
			childNode, err := sm.buildTreeRecursive(rootPath, secret.Path, maxDepth, currentDepth+1)
			if err != nil {
				sm.logger.Warnf("Failed to build tree for path '%s': %v", secret.Path, err)
				continue
			}
			if childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node, nil
}

// SearchSecrets searches for secrets by name pattern
func (sm *SecretsManager) SearchSecrets(pattern string, rootPath string) ([]*SecretNode, error) {
	var results []*SecretNode

	// This is a simple implementation - in a real scenario, you might want to
	// implement a more sophisticated search that walks the entire tree
	secrets, err := sm.ListSecrets(rootPath)
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets {
		if strings.Contains(strings.ToLower(secret.Name), strings.ToLower(pattern)) {
			if secret.IsSecret {
				// Get full secret data
				fullSecret, err := sm.GetSecret(secret.Path)
				if err != nil {
					sm.logger.Warnf("Failed to get secret '%s': %v", secret.Path, err)
					results = append(results, secret)
				} else {
					results = append(results, fullSecret)
				}
			} else {
				results = append(results, secret)
			}
		}
	}

	return results, nil
}

// AdvancedSearchOptions contains options for advanced search
type AdvancedSearchOptions struct {
	Pattern       string            `json:"pattern"`
	RootPath      string            `json:"root_path"`
	SearchType    SearchType        `json:"search_type"`
	KeyFilter     string            `json:"key_filter,omitempty"`
	ValueFilter   string            `json:"value_filter,omitempty"`
	MaxDepth      int               `json:"max_depth"`
	CaseSensitive bool              `json:"case_sensitive"`
	Regex         bool              `json:"regex"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// SearchType defines the type of search to perform
type SearchType int

const (
	SearchByName SearchType = iota
	SearchByPath
	SearchByKey
	SearchByValue
	SearchByMetadata
	SearchAll
)

// AdvancedSearch performs an advanced search with multiple criteria
func (sm *SecretsManager) AdvancedSearch(options *AdvancedSearchOptions) ([]*SearchResult, error) {
	var results []*SearchResult

	// Build search tree if needed
	rootNode, err := sm.BuildTree(options.RootPath, options.MaxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to build search tree: %w", err)
	}

	// Perform recursive search
	sm.searchRecursive(rootNode, options, &results)

	return results, nil
}

// SearchResult represents a search result with context
type SearchResult struct {
	Node       *SecretNode `json:"node"`
	MatchType  string      `json:"match_type"`
	MatchValue string      `json:"match_value"`
	Path       string      `json:"path"`
	Score      float64     `json:"score"`
}

// searchRecursive recursively searches through the tree
func (sm *SecretsManager) searchRecursive(node *SecretNode, options *AdvancedSearchOptions, results *[]*SearchResult) {
	if node == nil {
		return
	}

	// Check if this node matches the search criteria
	if sm.matchesSearchCriteria(node, options) {
		result := &SearchResult{
			Node:       node,
			Path:       node.Path,
			MatchType:  sm.getMatchType(node, options),
			MatchValue: sm.getMatchValue(node, options),
			Score:      sm.calculateScore(node, options),
		}
		*results = append(*results, result)
	}

	// Recursively search children
	for _, child := range node.Children {
		sm.searchRecursive(child, options, results)
	}
}

// matchesSearchCriteria checks if a node matches the search criteria
func (sm *SecretsManager) matchesSearchCriteria(node *SecretNode, options *AdvancedSearchOptions) bool {
	pattern := options.Pattern
	if !options.CaseSensitive {
		pattern = strings.ToLower(pattern)
	}

	switch options.SearchType {
	case SearchByName:
		return sm.matchesString(node.Name, pattern, options)
	case SearchByPath:
		return sm.matchesString(node.Path, pattern, options)
	case SearchByKey:
		return sm.matchesSecretKeys(node, pattern, options)
	case SearchByValue:
		return sm.matchesSecretValues(node, pattern, options)
	case SearchByMetadata:
		return sm.matchesMetadata(node, options)
	case SearchAll:
		return sm.matchesString(node.Name, pattern, options) ||
			sm.matchesString(node.Path, pattern, options) ||
			sm.matchesSecretKeys(node, pattern, options) ||
			sm.matchesSecretValues(node, pattern, options)
	default:
		return false
	}
}

// matchesString checks if a string matches the pattern
func (sm *SecretsManager) matchesString(text, pattern string, options *AdvancedSearchOptions) bool {
	if !options.CaseSensitive {
		text = strings.ToLower(text)
	}

	if options.Regex {
		// Simple regex matching - in a real implementation, you'd use regexp package
		return strings.Contains(text, pattern)
	}

	return strings.Contains(text, pattern)
}

// matchesSecretKeys checks if any secret key matches the pattern
func (sm *SecretsManager) matchesSecretKeys(node *SecretNode, pattern string, options *AdvancedSearchOptions) bool {
	if !node.IsSecret || node.Data == nil {
		return false
	}

	for key := range node.Data {
		if sm.matchesString(key, pattern, options) {
			return true
		}
	}

	return false
}

// matchesSecretValues checks if any secret value matches the pattern
func (sm *SecretsManager) matchesSecretValues(node *SecretNode, pattern string, options *AdvancedSearchOptions) bool {
	if !node.IsSecret || node.Data == nil {
		return false
	}

	for _, value := range node.Data {
		valueStr := fmt.Sprintf("%v", value)
		if sm.matchesString(valueStr, pattern, options) {
			return true
		}
	}

	return false
}

// matchesMetadata checks if metadata matches the criteria
func (sm *SecretsManager) matchesMetadata(node *SecretNode, options *AdvancedSearchOptions) bool {
	if node.Metadata == nil || len(options.Metadata) == 0 {
		return false
	}

	// This is a simplified implementation
	// In a real scenario, you'd check specific metadata fields
	return true
}

// getMatchType returns the type of match found
func (sm *SecretsManager) getMatchType(node *SecretNode, options *AdvancedSearchOptions) string {
	switch options.SearchType {
	case SearchByName:
		return "name"
	case SearchByPath:
		return "path"
	case SearchByKey:
		return "key"
	case SearchByValue:
		return "value"
	case SearchByMetadata:
		return "metadata"
	case SearchAll:
		return "multiple"
	default:
		return "unknown"
	}
}

// getMatchValue returns the value that matched
func (sm *SecretsManager) getMatchValue(node *SecretNode, options *AdvancedSearchOptions) string {
	switch options.SearchType {
	case SearchByName:
		return node.Name
	case SearchByPath:
		return node.Path
	case SearchByKey:
		// Return first matching key
		if node.Data != nil {
			for key := range node.Data {
				if sm.matchesString(key, options.Pattern, options) {
					return key
				}
			}
		}
		return ""
	case SearchByValue:
		// Return first matching value
		if node.Data != nil {
			for _, value := range node.Data {
				valueStr := fmt.Sprintf("%v", value)
				if sm.matchesString(valueStr, options.Pattern, options) {
					return valueStr
				}
			}
		}
		return ""
	default:
		return node.Name
	}
}

// calculateScore calculates a relevance score for the search result
func (sm *SecretsManager) calculateScore(node *SecretNode, options *AdvancedSearchOptions) float64 {
	score := 0.0
	pattern := options.Pattern
	if !options.CaseSensitive {
		pattern = strings.ToLower(pattern)
	}

	// Base score for exact matches
	if strings.EqualFold(node.Name, pattern) {
		score += 100.0
	} else if strings.HasPrefix(strings.ToLower(node.Name), pattern) {
		score += 50.0
	} else if strings.Contains(strings.ToLower(node.Name), pattern) {
		score += 25.0
	}

	// Bonus for path matches
	if strings.Contains(strings.ToLower(node.Path), pattern) {
		score += 10.0
	}

	// Bonus for secret data matches
	if node.IsSecret && node.Data != nil {
		for key, value := range node.Data {
			if strings.Contains(strings.ToLower(key), pattern) {
				score += 15.0
			}
			valueStr := fmt.Sprintf("%v", value)
			if strings.Contains(strings.ToLower(valueStr), pattern) {
				score += 5.0
			}
		}
	}

	return score
}

// SearchSecretsByValue searches for secrets containing specific values
func (sm *SecretsManager) SearchSecretsByValue(valuePattern string, rootPath string) ([]*SearchResult, error) {
	options := &AdvancedSearchOptions{
		Pattern:       valuePattern,
		RootPath:      rootPath,
		SearchType:    SearchByValue,
		MaxDepth:      10,
		CaseSensitive: false,
		Regex:         false,
	}

	return sm.AdvancedSearch(options)
}

// SearchSecretsByKey searches for secrets containing specific keys
func (sm *SecretsManager) SearchSecretsByKey(keyPattern string, rootPath string) ([]*SearchResult, error) {
	options := &AdvancedSearchOptions{
		Pattern:       keyPattern,
		RootPath:      rootPath,
		SearchType:    SearchByKey,
		MaxDepth:      10,
		CaseSensitive: false,
		Regex:         false,
	}

	return sm.AdvancedSearch(options)
}
