package backend

import (
	"path/filepath"
	"sync"

	"github.com/rvolykh/vui/internal/models"
)

type cacheEntry struct {
	secrets []*models.SecretNode
	secret  *models.SecretNode
}

type secretCache struct {
	cache map[string]cacheEntry
	mu    sync.RWMutex
}

func newSecretCache() *secretCache {
	return &secretCache{
		cache: make(map[string]cacheEntry),
	}
}

// normalizePath normalizes a path for consistent caching
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	// Clean the path and ensure consistent format
	normalized := filepath.Clean(path)
	if normalized == "." || normalized == "/" {
		return ""
	}
	return normalized
}

// GetListSecrets retrieves cached list of secrets, returns (secrets, found)
func (c *secretCache) GetListSecrets(path string) ([]*models.SecretNode, bool) {
	normalizedPath := normalizePath(path)

	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, found := c.cache[normalizedPath]; found && entry.secrets != nil {
		return entry.secrets, true
	}
	return nil, false
}

// GetSecret retrieves cached secret, returns (secret, found)
func (c *secretCache) GetSecret(path string) (*models.SecretNode, bool) {
	normalizedPath := normalizePath(path)

	c.mu.RLock()
	defer c.mu.RUnlock()

	if entry, found := c.cache[normalizedPath]; found && entry.secret != nil {
		return entry.secret, true
	}
	return nil, false
}

// SetListSecrets stores a list of secrets in the cache
func (c *secretCache) SetListSecrets(path string, secrets []*models.SecretNode) {
	normalizedPath := normalizePath(path)

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, found := c.cache[normalizedPath]; found {
		entry.secrets = secrets
		c.cache[normalizedPath] = entry
	} else {
		c.cache[normalizedPath] = cacheEntry{secrets: secrets}
	}
}

// SetSecret stores a secret in the cache
func (c *secretCache) SetSecret(path string, secret *models.SecretNode) {
	normalizedPath := normalizePath(path)

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, found := c.cache[normalizedPath]; found {
		entry.secret = secret
		c.cache[normalizedPath] = entry
	} else {
		c.cache[normalizedPath] = cacheEntry{secret: secret}
	}
}

// Invalidate invalidates cache entries for the given path and all parent paths
func (c *secretCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Normalize the path first
	normalizedPath := normalizePath(path)

	// Invalidate the path itself and all parent paths
	currentPath := normalizedPath
	for {
		delete(c.cache, currentPath)
		if currentPath == "" {
			break
		}
		currentPath = filepath.Dir(currentPath)
		// Normalize the parent path
		currentPath = normalizePath(currentPath)
	}
}
