package utils

import (
	"os"
)

func HomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	return homeDir
}
