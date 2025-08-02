package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/invoicer/storage"
	"github.com/user/invoicer/ui"
)

func main() {
	dataDir := "./data"
	store, err := storage.NewJSONStorage(dataDir)
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	
	p := tea.NewProgram(ui.NewMainMenuModel(store), tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}