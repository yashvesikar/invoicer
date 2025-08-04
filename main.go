package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/invoicer/config"
	"github.com/user/invoicer/storage"
	"github.com/user/invoicer/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// If no config exists, run setup
	if cfg == nil {
		cfg, err = config.RunSetup()
		if err != nil {
			log.Fatal("Setup failed:", err)
		}
		fmt.Println("\nConfiguration saved successfully!")
	}

	// Ensure directories exist
	if err := cfg.EnsureDirectories(); err != nil {
		log.Fatal("Failed to create directories:", err)
	}

	// Check for and offer data migration
	if err := config.CheckAndMigrate(cfg); err != nil {
		log.Fatal("Migration failed:", err)
	}

	// Initialize storage with configured data directory
	store, err := storage.NewJSONStorage(cfg.DataDir())
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	
	// Pass config to UI
	p := tea.NewProgram(ui.NewMainMenuModel(store, cfg), tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}