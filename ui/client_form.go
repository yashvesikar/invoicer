package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/user/invoicer/models"
)

type ClientFormModel struct {
	nameInput    textinput.Model
	emailInput   textinput.Model
	addressInput textinput.Model
	focusIndex   int
	inputs       []textinput.Model
	storage      models.Storage
	client       *models.Client
	isEdit       bool
	err          error
}

func NewClientFormModel(storage models.Storage, client *models.Client) ClientFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "John Doe"
	nameInput.Focus()
	
	emailInput := textinput.New()
	emailInput.Placeholder = "john@example.com"
	
	addressInput := textinput.New()
	addressInput.Placeholder = "123 Main St, City, Country"
	addressInput.Width = 50
	
	isEdit := client != nil
	if isEdit {
		nameInput.SetValue(client.Name)
		emailInput.SetValue(client.Email)
		addressInput.SetValue(client.Address)
	}
	
	return ClientFormModel{
		nameInput:    nameInput,
		emailInput:   emailInput,
		addressInput: addressInput,
		inputs:       []textinput.Model{nameInput, emailInput, addressInput},
		storage:      storage,
		client:       client,
		isEdit:       isEdit,
	}
}

func (m ClientFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ClientFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return NewClientListModel(m.storage), func() tea.Msg { return BackToClientListMsg{} }
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			
			if s == "enter" && m.focusIndex == len(m.inputs) {
				if err := m.saveClient(); err != nil {
					m.err = err
					return m, nil
				}
				return NewClientListModel(m.storage), func() tea.Msg { return BackToClientListMsg{} }
			}
			
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}
			
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = formInputStyle
					m.inputs[i].TextStyle = formInputStyle
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = dimStyle
					m.inputs[i].TextStyle = dimStyle
				}
			}
			
			return m, tea.Batch(cmds...)
		}
	}
	
	cmd := m.updateInputs(msg)
	
	return m, cmd
}

func (m *ClientFormModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	
	m.nameInput = m.inputs[0]
	m.emailInput = m.inputs[1]
	m.addressInput = m.inputs[2]
	
	return tea.Batch(cmds...)
}

func (m *ClientFormModel) saveClient() error {
	name := strings.TrimSpace(m.nameInput.Value())
	email := strings.TrimSpace(m.emailInput.Value())
	address := strings.TrimSpace(m.addressInput.Value())
	
	if name == "" {
		return fmt.Errorf("name is required")
	}
	
	if email == "" {
		return fmt.Errorf("email is required")
	}
	
	if m.isEdit {
		m.client.Update(name, address, email)
		return m.storage.UpdateClient(m.client)
	}
	
	client := models.NewClient(name, address, email)
	return m.storage.SaveClient(client)
}

func (m ClientFormModel) View() string {
	var s strings.Builder
	
	title := "New Client"
	if m.isEdit {
		title = "Edit Client"
	}
	s.WriteString(titleStyle.Render(title) + "\n\n")
	
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
	}
	
	labels := []string{"Name:", "Email:", "Address:"}
	for i, label := range labels {
		s.WriteString(formLabelStyle.Render(label))
		s.WriteString(m.inputs[i].View())
		if i < len(labels)-1 {
			s.WriteString("\n")
		}
	}
	
	s.WriteString("\n\n")
	
	button := "[ Save ]"
	if m.focusIndex == len(m.inputs) {
		button = selectedStyle.Render("[ Save ]")
	}
	s.WriteString(button)
	
	s.WriteString("\n\n" + helpStyle.Render("tab/shift+tab navigate • enter save • esc cancel"))
	
	return appStyle.Render(s.String())
}