package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/invoicer/backup"
	"github.com/user/invoicer/config"
	"github.com/user/invoicer/storage"
	"github.com/user/invoicer/ui"
)

func main() {
	var (
		backupFlag  = flag.Bool("backup", false, "Create a backup of all data")
		backupPath  = flag.String("backup-path", "", "Path to save backup (optional)")
		restoreFlag = flag.String("restore", "", "Restore from a backup file")
	)
	flag.Parse()

	if *backupFlag {
		if err := backup.CreateBackup(*backupPath); err != nil {
			log.Fatal("Backup failed:", err)
		}
		os.Exit(0)
	}

	if *restoreFlag != "" {
		if err := backup.RestoreBackup(*restoreFlag); err != nil {
			log.Fatal("Restore failed:", err)
		}
		os.Exit(0)
	}
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