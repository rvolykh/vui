package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rvolykh/vui/internal/app"
)

var (
	Version   = "unknown"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Check for --version flag
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		const link = "https://github.com/rvolykh/vui"
		fmt.Printf("%s / %s / %s / %s\n", Version, GitCommit, BuildTime, link)
		os.Exit(0)
	}

	// Create new application instance
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
		os.Exit(1)
	}
}
