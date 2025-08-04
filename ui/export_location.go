package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
)

type exportLocationMode int

const (
	exportLocationModeSelect exportLocationMode = iota
	exportLocationModeCustomPath
)

type ExportLocationModel struct {
	mode         exportLocationMode
	cursor       int
	customInput  textinput.Model
	selectedPath string
	invoiceNumber string
	err          error
}

func NewExportLocationModel(invoiceNumber string) ExportLocationModel {
	input := textinput.New()
	input.Placeholder = "Enter custom directory path..."
	input.Width = 50

	cwd, _ := os.Getwd()
	
	return ExportLocationModel{
		mode:          exportLocationModeSelect,
		customInput:   input,
		selectedPath:  cwd,
		invoiceNumber: invoiceNumber,
	}
}

func (m ExportLocationModel) Init() tea.Cmd {
	return nil
}

func (m ExportLocationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case exportLocationModeSelect:
		return m.updateSelect(msg)
	case exportLocationModeCustomPath:
		return m.updateCustomPath(msg)
	}
	return m, nil
}

func (m ExportLocationModel) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Cancel export
			return m, func() tea.Msg { return CancelExportMsg{} }
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
				// Current directory selected
				cwd, err := os.Getwd()
				if err != nil {
					m.err = err
					return m, nil
				}
				m.selectedPath = cwd
				return m, func() tea.Msg { return ExportLocationSelectedMsg{Path: cwd} }
			} else {
				// Custom directory
				m.mode = exportLocationModeCustomPath
				m.customInput.Focus()
				return m, textinput.Blink
			}
		}
	}
	return m, nil
}

func (m ExportLocationModel) updateCustomPath(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Go back to selection
			m.mode = exportLocationModeSelect
			m.customInput.Blur()
			return m, nil
		case "enter":
			path := strings.TrimSpace(m.customInput.Value())
			if path == "" {
				m.err = fmt.Errorf("path cannot be empty")
				return m, nil
			}
			
			// Expand ~ to home directory
			if strings.HasPrefix(path, "~/") {
				home, err := os.UserHomeDir()
				if err != nil {
					m.err = err
					return m, nil
				}
				path = filepath.Join(home, path[2:])
			}
			
			// Verify directory exists
			info, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					m.err = fmt.Errorf("directory does not exist: %s", path)
				} else {
					m.err = err
				}
				return m, nil
			}
			
			if !info.IsDir() {
				m.err = fmt.Errorf("path is not a directory: %s", path)
				return m, nil
			}
			
			m.selectedPath = path
			return m, func() tea.Msg { return ExportLocationSelectedMsg{Path: path} }
		}
	}

	// Update text input
	var cmd tea.Cmd
	m.customInput, cmd = m.customInput.Update(msg)
	return m, cmd
}

func (m ExportLocationModel) View() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Export Invoice to PDF") + "\n")
	s.WriteString(strings.Repeat("─", 40) + "\n")
	s.WriteString("Where would you like to save the PDF?\n\n")

	switch m.mode {
	case exportLocationModeSelect:
		// Show options
		cwd, _ := os.Getwd()
		options := []string{
			fmt.Sprintf("Current directory (%s)", cwd),
			"Custom directory...",
		}
		
		for i, option := range options {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
				s.WriteString(selectedListItemStyle.Render(cursor + option) + "\n")
			} else {
				s.WriteString(listItemStyle.Render(cursor + option) + "\n")
			}
		}
		
		s.WriteString("\n")
		fileName := fmt.Sprintf("invoice_%s.pdf", m.invoiceNumber)
		previewPath := filepath.Join(m.selectedPath, fileName)
		s.WriteString(helpStyle.Render(fmt.Sprintf("File will be saved as: %s", previewPath)) + "\n")
		
		if m.err != nil {
			s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		}
		
		s.WriteString("\n" + helpStyle.Render("↑/k up • ↓/j down • enter select • esc cancel"))
		
	case exportLocationModeCustomPath:
		s.WriteString("Enter the directory path:\n\n")
		s.WriteString(m.customInput.View() + "\n\n")
		
		if m.customInput.Value() != "" {
			path := strings.TrimSpace(m.customInput.Value())
			if strings.HasPrefix(path, "~/") {
				home, _ := os.UserHomeDir()
				path = filepath.Join(home, path[2:])
			}
			fileName := fmt.Sprintf("invoice_%s.pdf", m.invoiceNumber)
			previewPath := filepath.Join(path, fileName)
			s.WriteString(helpStyle.Render(fmt.Sprintf("File will be saved as: %s", previewPath)) + "\n")
		}
		
		if m.err != nil {
			s.WriteString("\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		}
		
		s.WriteString("\n" + helpStyle.Render("enter confirm • esc back"))
	}
	
	return appStyle.Render(s.String())
}

// Messages
type ExportLocationSelectedMsg struct {
	Path string
}

type CancelExportMsg struct{}