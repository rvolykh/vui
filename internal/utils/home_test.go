package utils

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHomeDir_Positive(t *testing.T) {
	home := HomeDir()
	require.NotEmpty(t, home)

	info, err := os.Stat(home)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestHomeDir_Negative(t *testing.T) {
	originalEnv := os.Environ()
	t.Cleanup(func() {
		// Clear the environment
		os.Clearenv()
		// Restore the original environment variables
		for _, envVar := range originalEnv {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	})

	os.Clearenv()

	home := HomeDir()
	require.Equal(t, ".", home)
}
