package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type migrationModel struct {
	config       *Config
	oldDataPath  string
	hasOldData   bool
	confirmMigrate bool
	cursor       int
	done         bool
	err          error
}

func CheckAndMigrate(cfg *Config) error {
	// Check multiple old data locations
	homeDir, _ := os.UserHomeDir()
	oldLocations := []string{
		"./data",                               // Original location
		filepath.Join(homeDir, ".invoicer"),    // Old default location
	}
	
	var oldDataPath string
	hasOldData := false
	
	for _, location := range oldLocations {
		if location == cfg.DataPath {
			// Skip if this is our current location
			continue
		}
		
		dataDir := location
		if location != "./data" {
			dataDir = filepath.Join(location, "data")
		}
		
		if info, err := os.Stat(dataDir); err == nil && info.IsDir() {
			// Check if there are JSON files
			files, _ := os.ReadDir(dataDir)
			for _, file := range files {
				if filepath.Ext(file.Name()) == ".json" {
					hasOldData = true
					oldDataPath = location
					break
				}
			}
		}
		
		if hasOldData {
			break
		}
	}
	
	if !hasOldData {
		return nil
	}
	
	// Run migration UI
	m := migrationModel{
		config:      cfg,
		oldDataPath: oldDataPath,
		hasOldData:  true,
	}
	
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	
	final := finalModel.(migrationModel)
	return final.err
}

func (m migrationModel) Init() tea.Cmd {
	return nil
}

func (m migrationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
			
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			
		case "down", "j":
			if m.cursor < 1 {
				m.cursor++
			}
			
		case "enter":
			if m.cursor == 0 {
				// Migrate data
				if err := m.migrateData(); err != nil {
					m.err = err
					m.done = true
					return m, tea.Quit
				}
				m.done = true
				return m, tea.Quit
			} else {
				// Skip migration
				m.done = true
				return m, tea.Quit
			}
		}
	}
	
	return m, nil
}

func (m migrationModel) View() string {
	if m.done {
		if m.err != nil {
			return migrationErrorStyle.Render(fmt.Sprintf("Migration failed: %v", m.err))
		}
		if m.cursor == 0 {
			return successStyle.Render("✓ Data migration completed successfully!")
		}
		return ""
	}
	
	var s string
	
	s += migrationTitleStyle.Render("Data Migration Detected") + "\n\n"
	s += "Found existing data in: " + highlightStyle.Render(m.oldDataPath) + "\n"
	s += "New data location: " + highlightStyle.Render(m.config.DataDir()) + "\n\n"
	s += "Would you like to migrate your existing data to the new location?\n\n"
	
	options := []string{
		"Yes, migrate my data",
		"No, keep data in current location",
	}
	
	for i, option := range options {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			s += migrationSelectedStyle.Render(cursor + option) + "\n"
		} else {
			s += migrationListStyle.Render(cursor + option) + "\n"
		}
	}
	
	s += "\n" + migrationHelpStyle.Render("↑/k up • ↓/j down • enter select • esc cancel")
	
	return s
}

func (m *migrationModel) migrateData() error {
	// Create destination directories
	if err := os.MkdirAll(m.config.DataDir(), 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	if err := os.MkdirAll(m.config.TemplatesDir(), 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}
	
	// Determine source directories based on old location
	var dataDir, templatesDir string
	if m.oldDataPath == "./data" {
		// Original structure: ./data and ./templates
		dataDir = "./data"
		templatesDir = "./templates"
	} else {
		// New structure: everything under one directory
		dataDir = filepath.Join(m.oldDataPath, "data")
		templatesDir = filepath.Join(m.oldDataPath, "templates")
	}
	
	// Copy data files
	if info, err := os.Stat(dataDir); err == nil && info.IsDir() {
		files, err := os.ReadDir(dataDir)
		if err != nil {
			return fmt.Errorf("failed to read data directory: %w", err)
		}
		
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".json" {
				src := filepath.Join(dataDir, file.Name())
				dst := filepath.Join(m.config.DataDir(), file.Name())
				
				if err := copyFile(src, dst); err != nil {
					return fmt.Errorf("failed to copy %s: %w", file.Name(), err)
				}
			}
		}
	}
	
	// Copy templates
	if info, err := os.Stat(templatesDir); err == nil && info.IsDir() {
		if err := copyDir(templatesDir, m.config.TemplatesDir()); err != nil {
			return fmt.Errorf("failed to copy templates: %w", err)
		}
	}
	
	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()
	
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	
	_, err = io.Copy(destination, source)
	return err
}

func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	
	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	
	return nil
}

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	migrationErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	migrationTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	migrationHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	migrationSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	migrationListStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)