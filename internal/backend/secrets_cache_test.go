package backend

import (
	"sync"
	"testing"

	"github.com/rvolykh/vui/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretCache(t *testing.T) {
	cache := newSecretCache()

	assert.NotNil(t, cache)
	assert.NotNil(t, cache.cache)
	assert.Equal(t, 0, len(cache.cache))
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "simple path",
			input:    "path/to/secret",
			expected: "path/to/secret",
		},
		{
			name:     "path with dot",
			input:    ".",
			expected: "",
		},
		{
			name:     "path with slash",
			input:    "/",
			expected: "",
		},
		{
			name:     "path with trailing slash",
			input:    "path/to/",
			expected: "path/to",
		},
		{
			name:     "path with double slashes",
			input:    "path//to//secret",
			expected: "path/to/secret",
		},
		{
			name:     "path with dot segments",
			input:    "path/./to/../secret",
			expected: "path/secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecretCache_GetListSecrets(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		cache := newSecretCache()
		secrets := []*models.SecretNode{
			{Name: "secret1", Path: "secret1", IsSecret: true},
			{Name: "secret2", Path: "secret2", IsSecret: true},
		}
		cache.SetListSecrets("path/to", secrets)

		result, found := cache.GetListSecrets("path/to")

		assert.True(t, found)
		assert.Equal(t, secrets, result)
	})

	t.Run("cache miss", func(t *testing.T) {
		cache := newSecretCache()

		result, found := cache.GetListSecrets("path/to")

		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("cache miss with nil secrets", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}
		cache.SetSecret("path/to/secret", secret)

		result, found := cache.GetListSecrets("path/to/secret")

		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("path normalization", func(t *testing.T) {
		cache := newSecretCache()
		secrets := []*models.SecretNode{{Name: "secret", Path: "secret", IsSecret: true}}
		cache.SetListSecrets("path/to", secrets)

		result, found := cache.GetListSecrets("path//to")

		assert.True(t, found)
		assert.Equal(t, secrets, result)
	})

	t.Run("empty path", func(t *testing.T) {
		cache := newSecretCache()
		secrets := []*models.SecretNode{{Name: "secret", Path: "secret", IsSecret: true}}
		cache.SetListSecrets("", secrets)

		result, found := cache.GetListSecrets("")

		assert.True(t, found)
		assert.Equal(t, secrets, result)
	})
}

func TestSecretCache_GetSecret(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}
		cache.SetSecret("path/to/secret", secret)

		result, found := cache.GetSecret("path/to/secret")

		assert.True(t, found)
		assert.Equal(t, secret, result)
	})

	t.Run("cache miss", func(t *testing.T) {
		cache := newSecretCache()

		result, found := cache.GetSecret("path/to/secret")

		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("cache miss with nil secret", func(t *testing.T) {
		cache := newSecretCache()
		secrets := []*models.SecretNode{{Name: "secret", Path: "path/to/secret", IsSecret: true}}
		cache.SetListSecrets("path/to/secret", secrets)

		result, found := cache.GetSecret("path/to/secret")

		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("path normalization", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}
		cache.SetSecret("path/to/secret", secret)

		result, found := cache.GetSecret("path//to//secret")

		assert.True(t, found)
		assert.Equal(t, secret, result)
	})

	t.Run("empty path", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "", IsSecret: true}
		cache.SetSecret("", secret)

		result, found := cache.GetSecret("")

		assert.True(t, found)
		assert.Equal(t, secret, result)
	})
}

func TestSecretCache_SetListSecrets(t *testing.T) {
	t.Run("set new entry", func(t *testing.T) {
		cache := newSecretCache()
		secrets := []*models.SecretNode{{Name: "secret", Path: "secret", IsSecret: true}}

		cache.SetListSecrets("path/to", secrets)

		result, found := cache.GetListSecrets("path/to")
		assert.True(t, found)
		assert.Equal(t, secrets, result)
	})

	t.Run("update existing entry", func(t *testing.T) {
		cache := newSecretCache()
		secrets1 := []*models.SecretNode{{Name: "secret1", Path: "secret1", IsSecret: true}}
		secrets2 := []*models.SecretNode{{Name: "secret2", Path: "secret2", IsSecret: true}}

		cache.SetListSecrets("path/to", secrets1)
		cache.SetListSecrets("path/to", secrets2)

		result, found := cache.GetListSecrets("path/to")
		assert.True(t, found)
		assert.Equal(t, secrets2, result)
	})

	t.Run("preserve secret when setting list", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}
		secrets := []*models.SecretNode{{Name: "list", Path: "list", IsSecret: true}}

		cache.SetSecret("path/to/secret", secret)
		cache.SetListSecrets("path/to/secret", secrets)

		// Should still get the secret
		result, found := cache.GetSecret("path/to/secret")
		assert.True(t, found)
		assert.Equal(t, secret, result)

		// Should also get the list
		listResult, found := cache.GetListSecrets("path/to/secret")
		assert.True(t, found)
		assert.Equal(t, secrets, listResult)
	})

	t.Run("empty path", func(t *testing.T) {
		cache := newSecretCache()
		secrets := []*models.SecretNode{{Name: "secret", Path: "secret", IsSecret: true}}

		cache.SetListSecrets("", secrets)

		result, found := cache.GetListSecrets("")
		assert.True(t, found)
		assert.Equal(t, secrets, result)
	})
}

func TestSecretCache_SetSecret(t *testing.T) {
	t.Run("set new entry", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}

		cache.SetSecret("path/to/secret", secret)

		result, found := cache.GetSecret("path/to/secret")
		assert.True(t, found)
		assert.Equal(t, secret, result)
	})

	t.Run("update existing entry", func(t *testing.T) {
		cache := newSecretCache()
		secret1 := &models.SecretNode{Name: "secret1", Path: "path/to/secret", IsSecret: true}
		secret2 := &models.SecretNode{Name: "secret2", Path: "path/to/secret", IsSecret: true}

		cache.SetSecret("path/to/secret", secret1)
		cache.SetSecret("path/to/secret", secret2)

		result, found := cache.GetSecret("path/to/secret")
		assert.True(t, found)
		assert.Equal(t, secret2, result)
	})

	t.Run("preserve list when setting secret", func(t *testing.T) {
		cache := newSecretCache()
		secrets := []*models.SecretNode{{Name: "list", Path: "list", IsSecret: true}}
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}

		cache.SetListSecrets("path/to/secret", secrets)
		cache.SetSecret("path/to/secret", secret)

		// Should still get the list
		result, found := cache.GetListSecrets("path/to/secret")
		assert.True(t, found)
		assert.Equal(t, secrets, result)

		// Should also get the secret
		secretResult, found := cache.GetSecret("path/to/secret")
		assert.True(t, found)
		assert.Equal(t, secret, secretResult)
	})

	t.Run("empty path", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "", IsSecret: true}

		cache.SetSecret("", secret)

		result, found := cache.GetSecret("")
		assert.True(t, found)
		assert.Equal(t, secret, result)
	})
}

func TestSecretCache_Invalidate(t *testing.T) {
	t.Run("invalidate single path", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}
		cache.SetSecret("path/to/secret", secret)

		cache.Invalidate("path/to/secret")

		result, found := cache.GetSecret("path/to/secret")
		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("invalidate parent paths", func(t *testing.T) {
		cache := newSecretCache()
		secret1 := &models.SecretNode{Name: "secret1", Path: "path/to/secret1", IsSecret: true}
		secret2 := &models.SecretNode{Name: "secret2", Path: "path/to/secret2", IsSecret: true}
		list1 := []*models.SecretNode{{Name: "list", Path: "path/to", IsSecret: false}}
		list2 := []*models.SecretNode{{Name: "list", Path: "path", IsSecret: false}}

		cache.SetSecret("path/to/secret1", secret1)
		cache.SetSecret("path/to/secret2", secret2)
		cache.SetListSecrets("path/to", list1)
		cache.SetListSecrets("path", list2)

		cache.Invalidate("path/to/secret1")

		// The secret itself should be invalidated
		result, found := cache.GetSecret("path/to/secret1")
		assert.False(t, found)
		assert.Nil(t, result)

		// Parent path should be invalidated
		resultList, found := cache.GetListSecrets("path/to")
		assert.False(t, found)
		assert.Nil(t, resultList)

		// Grandparent path should be invalidated
		resultList2, found := cache.GetListSecrets("path")
		assert.False(t, found)
		assert.Nil(t, resultList2)

		// Root should be invalidated
		resultRoot, found := cache.GetListSecrets("")
		assert.False(t, found)
		assert.Nil(t, resultRoot)

		// Sibling secret should still be cached (only parent paths are invalidated, not siblings)
		// Note: This is the expected behavior - invalidating a path removes it and its ancestors,
		// but not sibling paths at the same level
		result2, found := cache.GetSecret("path/to/secret2")
		assert.True(t, found)
		assert.Equal(t, secret2, result2)
	})

	t.Run("invalidate root path", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}
		list := []*models.SecretNode{{Name: "list", Path: "path", IsSecret: false}}
		cache.SetSecret("path/to/secret", secret)
		cache.SetListSecrets("path", list)
		cache.SetListSecrets("", []*models.SecretNode{})

		cache.Invalidate("")

		result, found := cache.GetListSecrets("")
		assert.False(t, found)
		assert.Nil(t, result)

		// Child paths should still be cached (only root was invalidated)
		resultSecret, found := cache.GetSecret("path/to/secret")
		assert.True(t, found)
		assert.Equal(t, secret, resultSecret)
	})

	t.Run("invalidate with path normalization", func(t *testing.T) {
		cache := newSecretCache()
		secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}
		cache.SetSecret("path/to/secret", secret)

		cache.Invalidate("path//to//secret")

		result, found := cache.GetSecret("path/to/secret")
		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("invalidate empty path", func(t *testing.T) {
		cache := newSecretCache()
		list := []*models.SecretNode{{Name: "list", Path: "", IsSecret: false}}
		cache.SetListSecrets("", list)

		cache.Invalidate("")

		result, found := cache.GetListSecrets("")
		assert.False(t, found)
		assert.Nil(t, result)
	})
}

func TestSecretCache_ConcurrentAccess(t *testing.T) {
	cache := newSecretCache()
	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup

	// Concurrent writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				secret := &models.SecretNode{
					Name:     "secret",
					Path:     "path/to/secret",
					IsSecret: true,
				}
				cache.SetSecret("path/to/secret", secret)
				cache.GetSecret("path/to/secret")
				cache.SetListSecrets("path/to", []*models.SecretNode{secret})
				cache.GetListSecrets("path/to")
			}
		}(i)
	}

	// Concurrent invalidations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				cache.Invalidate("path/to/secret")
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is still functional
	secret := &models.SecretNode{Name: "final", Path: "final", IsSecret: true}
	cache.SetSecret("final", secret)
	result, found := cache.GetSecret("final")
	assert.True(t, found)
	assert.Equal(t, secret, result)
}

func TestSecretCache_CombinedOperations(t *testing.T) {
	cache := newSecretCache()

	// Set both list and secret for the same path
	secrets := []*models.SecretNode{
		{Name: "secret1", Path: "path/to/secret1", IsSecret: true},
		{Name: "secret2", Path: "path/to/secret2", IsSecret: true},
	}
	secret := &models.SecretNode{Name: "secret", Path: "path/to/secret", IsSecret: true}

	cache.SetListSecrets("path/to", secrets)
	cache.SetSecret("path/to/secret", secret)

	// Both should be retrievable
	listResult, found := cache.GetListSecrets("path/to")
	assert.True(t, found)
	assert.Equal(t, secrets, listResult)

	secretResult, found := cache.GetSecret("path/to/secret")
	assert.True(t, found)
	assert.Equal(t, secret, secretResult)

	// Invalidate should remove both
	cache.Invalidate("path/to/secret")

	listResult2, found := cache.GetListSecrets("path/to")
	assert.False(t, found)
	assert.Nil(t, listResult2)

	secretResult2, found := cache.GetSecret("path/to/secret")
	assert.False(t, found)
	assert.Nil(t, secretResult2)
}
