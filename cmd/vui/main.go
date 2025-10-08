package main

import (
	"log"
	"os"

	"github.com/rvolykh/vui/internal/app"
)

func main() {
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
