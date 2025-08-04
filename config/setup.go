package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type setupModel struct {
	config     *Config
	inputs     []textinput.Model
	focusIndex int
	err        error
	done       bool
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func RunSetup() (*Config, error) {
	config := DefaultConfig()
	
	m := setupModel{
		config: config,
		inputs: make([]textinput.Model, 9),
	}

	// Data path input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = config.DataPath
	m.inputs[0].Focus()
	m.inputs[0].CharLimit = 256
	m.inputs[0].Width = 50
	m.inputs[0].Prompt = "Data directory: "

	// Company name input
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = config.CompanyName
	m.inputs[1].CharLimit = 100
	m.inputs[1].Width = 50
	m.inputs[1].Prompt = "Company name: "

	// Company address input
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = config.CompanyAddress
	m.inputs[2].CharLimit = 200
	m.inputs[2].Width = 50
	m.inputs[2].Prompt = "Company address: "

	// Company email input
	m.inputs[3] = textinput.New()
	m.inputs[3].Placeholder = config.CompanyEmail
	m.inputs[3].CharLimit = 100
	m.inputs[3].Width = 50
	m.inputs[3].Prompt = "Company email: "

	// Zelle account input
	m.inputs[4] = textinput.New()
	m.inputs[4].Placeholder = "email@example.com or phone number (optional)"
	m.inputs[4].CharLimit = 100
	m.inputs[4].Width = 50
	m.inputs[4].Prompt = "Zelle account: "

	// Venmo account input
	m.inputs[5] = textinput.New()
	m.inputs[5].Placeholder = "@username (optional)"
	m.inputs[5].CharLimit = 50
	m.inputs[5].Width = 50
	m.inputs[5].Prompt = "Venmo account: "

	// Bank name input
	m.inputs[6] = textinput.New()
	m.inputs[6].Placeholder = "Bank name (optional)"
	m.inputs[6].CharLimit = 100
	m.inputs[6].Width = 50
	m.inputs[6].Prompt = "Bank name: "

	// Bank routing input
	m.inputs[7] = textinput.New()
	m.inputs[7].Placeholder = "9-digit routing number (optional)"
	m.inputs[7].CharLimit = 9
	m.inputs[7].Width = 50
	m.inputs[7].Prompt = "Bank routing: "

	// Bank account input
	m.inputs[8] = textinput.New()
	m.inputs[8].Placeholder = "Account number (optional)"
	m.inputs[8].CharLimit = 20
	m.inputs[8].Width = 50
	m.inputs[8].Prompt = "Bank account: "

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := finalModel.(setupModel)
	if final.err != nil {
		return nil, final.err
	}

	return final.config, nil
}

func (m setupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.err = fmt.Errorf("setup cancelled")
			return m, tea.Quit

		case "tab", "down":
			m.focusIndex++
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}
			return m, m.updateFocus()

		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			return m, m.updateFocus()

		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				// Last field, save and exit
				if err := m.saveConfig(); err != nil {
					m.err = err
					return m, tea.Quit
				}
				m.done = true
				return m, tea.Quit
			}
			// Move to next field
			m.focusIndex++
			return m, m.updateFocus()
		}
	}

	// Update the focused input
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *setupModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i < len(m.inputs); i++ {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = blurredStyle
			m.inputs[i].TextStyle = blurredStyle
		}
	}
	return tea.Batch(cmds...)
}

func (m *setupModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m *setupModel) saveConfig() error {
	// Get values from inputs
	val := strings.TrimSpace(m.inputs[0].Value())
	if val != "" {
		// Expand ~ to home directory
		if strings.HasPrefix(val, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			val = filepath.Join(home, val[2:])
		}
		m.config.DataPath = val
	}

	// For required fields, only update if not empty
	if val := strings.TrimSpace(m.inputs[1].Value()); val != "" {
		m.config.CompanyName = val
	}

	if val := strings.TrimSpace(m.inputs[2].Value()); val != "" {
		m.config.CompanyAddress = val
	}

	if val := strings.TrimSpace(m.inputs[3].Value()); val != "" {
		m.config.CompanyEmail = val
	}

	// Payment information - always update (can be cleared)
	m.config.ZelleAccount = strings.TrimSpace(m.inputs[4].Value())
	m.config.VenmoAccount = strings.TrimSpace(m.inputs[5].Value())
	m.config.BankName = strings.TrimSpace(m.inputs[6].Value())
	m.config.BankRouting = strings.TrimSpace(m.inputs[7].Value())
	m.config.BankAccount = strings.TrimSpace(m.inputs[8].Value())

	// Create directories
	if err := m.config.EnsureDirectories(); err != nil {
		return err
	}

	// Save config
	return m.config.Save()
}

func (m setupModel) View() string {
	if m.done {
		return titleStyle.Render("✓ Setup complete!") + "\n"
	}

	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Welcome to Invoicer!") + "\n")
	s.WriteString("Let's set up your configuration.\n\n")

	// Company information section
	s.WriteString(blurredStyle.Render("Company Information") + "\n")
	for i := 0; i < 4; i++ {
		s.WriteString(m.inputs[i].View())
		s.WriteString("\n")
	}

	// Payment information section
	s.WriteString("\n" + blurredStyle.Render("Payment Information (optional)") + "\n")
	for i := 4; i < len(m.inputs); i++ {
		s.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			s.WriteString("\n")
		}
	}

	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("tab/shift+tab to navigate • enter to confirm • esc to cancel"))

	return s.String()
}

// NewSettingsEditor creates a setup model for editing existing configuration
func NewSettingsEditor(cfg *Config) tea.Model {
	m := setupModel{
		config: cfg,
		inputs: make([]textinput.Model, 9),
	}

	// Data path input
	m.inputs[0] = textinput.New()
	m.inputs[0].SetValue(cfg.DataPath)
	m.inputs[0].Focus()
	m.inputs[0].CharLimit = 256
	m.inputs[0].Width = 50
	m.inputs[0].Prompt = "Data directory: "

	// Company name input
	m.inputs[1] = textinput.New()
	m.inputs[1].SetValue(cfg.CompanyName)
	m.inputs[1].CharLimit = 100
	m.inputs[1].Width = 50
	m.inputs[1].Prompt = "Company name: "

	// Company address input
	m.inputs[2] = textinput.New()
	m.inputs[2].SetValue(cfg.CompanyAddress)
	m.inputs[2].CharLimit = 200
	m.inputs[2].Width = 50
	m.inputs[2].Prompt = "Company address: "

	// Company email input
	m.inputs[3] = textinput.New()
	m.inputs[3].SetValue(cfg.CompanyEmail)
	m.inputs[3].CharLimit = 100
	m.inputs[3].Width = 50
	m.inputs[3].Prompt = "Company email: "

	// Zelle account input
	m.inputs[4] = textinput.New()
	m.inputs[4].SetValue(cfg.ZelleAccount)
	m.inputs[4].Placeholder = "email@example.com or phone number (optional)"
	m.inputs[4].CharLimit = 100
	m.inputs[4].Width = 50
	m.inputs[4].Prompt = "Zelle account: "

	// Venmo account input
	m.inputs[5] = textinput.New()
	m.inputs[5].SetValue(cfg.VenmoAccount)
	m.inputs[5].Placeholder = "@username (optional)"
	m.inputs[5].CharLimit = 50
	m.inputs[5].Width = 50
	m.inputs[5].Prompt = "Venmo account: "

	// Bank name input
	m.inputs[6] = textinput.New()
	m.inputs[6].SetValue(cfg.BankName)
	m.inputs[6].Placeholder = "Bank name (optional)"
	m.inputs[6].CharLimit = 100
	m.inputs[6].Width = 50
	m.inputs[6].Prompt = "Bank name: "

	// Bank routing input
	m.inputs[7] = textinput.New()
	m.inputs[7].SetValue(cfg.BankRouting)
	m.inputs[7].Placeholder = "9-digit routing number (optional)"
	m.inputs[7].CharLimit = 9
	m.inputs[7].Width = 50
	m.inputs[7].Prompt = "Bank routing: "

	// Bank account input
	m.inputs[8] = textinput.New()
	m.inputs[8].SetValue(cfg.BankAccount)
	m.inputs[8].Placeholder = "Account number (optional)"
	m.inputs[8].CharLimit = 20
	m.inputs[8].Width = 50
	m.inputs[8].Prompt = "Bank account: "

	return settingsEditorModel{setupModel: m}
}

// settingsEditorModel wraps setupModel for settings editing
type settingsEditorModel struct {
	setupModel setupModel
}

func (m settingsEditorModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m settingsEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return SettingsCancelledMsg{} }

		case "tab", "down":
			m.setupModel.focusIndex++
			if m.setupModel.focusIndex >= len(m.setupModel.inputs) {
				m.setupModel.focusIndex = 0
			}
			return m, m.setupModel.updateFocus()

		case "shift+tab", "up":
			m.setupModel.focusIndex--
			if m.setupModel.focusIndex < 0 {
				m.setupModel.focusIndex = len(m.setupModel.inputs) - 1
			}
			return m, m.setupModel.updateFocus()

		case "enter":
			if m.setupModel.focusIndex == len(m.setupModel.inputs)-1 {
				// Last field, save and exit
				if err := m.setupModel.saveConfig(); err != nil {
					return m, func() tea.Msg { return SettingsErrorMsg{Err: err} }
				}
				return m, func() tea.Msg { return SettingsSavedMsg{} }
			}
			// Move to next field
			m.setupModel.focusIndex++
			return m, m.setupModel.updateFocus()
		}
	}

	// Update the focused input
	cmd := m.setupModel.updateInputs(msg)
	return m, cmd
}

func (m settingsEditorModel) View() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Settings") + "\n")
	s.WriteString("Edit your configuration.\n\n")

	// Company information section
	s.WriteString(blurredStyle.Render("Company Information") + "\n")
	for i := 0; i < 4; i++ {
		s.WriteString(m.setupModel.inputs[i].View())
		s.WriteString("\n")
	}

	// Payment information section
	s.WriteString("\n" + blurredStyle.Render("Payment Information (optional)") + "\n")
	for i := 4; i < len(m.setupModel.inputs); i++ {
		s.WriteString(m.setupModel.inputs[i].View())
		if i < len(m.setupModel.inputs)-1 {
			s.WriteString("\n")
		}
	}

	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("tab/shift+tab to navigate • enter to save • esc to cancel"))

	return s.String()
}

// Message types for settings editor
type SettingsSavedMsg struct{}
type SettingsCancelledMsg struct{}
type SettingsErrorMsg struct {
	Err error
}