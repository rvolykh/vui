package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoalesce(t *testing.T) {
	t.Run("test int first", func(t *testing.T) {
		actual := Coalesce(1, 2, 3)
		assert.Equal(t, 1, actual)
	})

	t.Run("test int second", func(t *testing.T) {
		actual := Coalesce(0, 2, 3)
		assert.Equal(t, 2, actual)
	})

	t.Run("test str first", func(t *testing.T) {
		actual := Coalesce("a", "", "c")
		assert.Equal(t, "a", actual)
	})

	t.Run("test str third", func(t *testing.T) {
		actual := Coalesce("", "", "c")
		assert.Equal(t, "c", actual)
	})
}
