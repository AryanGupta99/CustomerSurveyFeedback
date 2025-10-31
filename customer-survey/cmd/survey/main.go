package main

import (
	"customer-survey/internal/ui"
	"log"
)

func main() {
	// Launch native Windows desktop UI
	if err := ui.RunDesktopUI(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
