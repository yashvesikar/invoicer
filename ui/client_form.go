package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/shopspring/decimal"
	"github.com/user/invoicer/models"
)

type clientFormMode int

const (
	clientFormModeEdit clientFormMode = iota
	clientFormModeManageEmails
)

type ClientFormModel struct {
	nameInput       textinput.Model
	addressInput    textinput.Model
	hourlyRateInput textinput.Model
	emailInputs     []textinput.Model
	emails          []string
	focusIndex      int
	emailFocusIndex int
	storage         models.Storage
	client          *models.Client
	isEdit          bool
	mode            clientFormMode
	err             error
}

func NewClientFormModel(storage models.Storage, client *models.Client) ClientFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "John Doe"
	nameInput.Focus()
	
	addressInput := textinput.New()
	addressInput.Placeholder = "123 Main St, City, Country"
	addressInput.Width = 50
	
	hourlyRateInput := textinput.New()
	hourlyRateInput.Placeholder = "150.00"
	hourlyRateInput.Width = 15
	
	emails := []string{""}
	emailInputs := []textinput.Model{createEmailInput()}
	
	isEdit := client != nil
	if isEdit {
		nameInput.SetValue(client.Name)
		addressInput.SetValue(client.Address)
		hourlyRateInput.SetValue(client.DefaultHourlyRate.String())
		if len(client.Emails) > 0 {
			emails = client.Emails
			emailInputs = make([]textinput.Model, len(emails))
			for i, email := range emails {
				emailInputs[i] = createEmailInput()
				emailInputs[i].SetValue(email)
			}
		}
	}
	
	return ClientFormModel{
		nameInput:       nameInput,
		addressInput:    addressInput,
		hourlyRateInput: hourlyRateInput,
		emailInputs:     emailInputs,
		emails:          emails,
		storage:         storage,
		client:          client,
		isEdit:          isEdit,
		mode:            clientFormModeEdit,
	}
}

func createEmailInput() textinput.Model {
	input := textinput.New()
	input.Placeholder = "john@example.com"
	input.Width = 40
	return input
}

func (m ClientFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ClientFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case clientFormModeManageEmails:
		return m.updateManageEmails(msg)
	default:
		return m.updateEditMode(msg)
	}
}

func (m ClientFormModel) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return NewClientListModel(m.storage), func() tea.Msg { return BackToClientListMsg{} }
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			
			// Total fields: name, address, hourly rate, manage emails button, save button
			totalFields := 5
			
			if s == "enter" && m.focusIndex == 3 {
				// Enter manage emails mode
				m.mode = clientFormModeManageEmails
				m.emailFocusIndex = 0
				return m.updateEmailFocus()
			}
			
			if s == "enter" && m.focusIndex == totalFields-1 {
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
			
			if m.focusIndex >= totalFields {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = totalFields - 1
			}
			
			return m.updateMainFocus()
		}
	}
	
	cmd := m.updateMainInputs(msg)
	return m, cmd
}

func (m ClientFormModel) updateManageEmails(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = clientFormModeEdit
			// Save email values
			for i, input := range m.emailInputs {
				if i < len(m.emails) {
					m.emails[i] = strings.TrimSpace(input.Value())
				}
			}
			return m.updateMainFocus()
		case "+":
			// Add new email
			m.emails = append(m.emails, "")
			m.emailInputs = append(m.emailInputs, createEmailInput())
			m.emailFocusIndex = len(m.emailInputs) - 1
			return m.updateEmailFocus()
		case "-":
			// Remove current email if there's more than one
			if len(m.emails) > 1 && m.emailFocusIndex < len(m.emails) {
				m.emails = append(m.emails[:m.emailFocusIndex], m.emails[m.emailFocusIndex+1:]...)
				m.emailInputs = append(m.emailInputs[:m.emailFocusIndex], m.emailInputs[m.emailFocusIndex+1:]...)
				if m.emailFocusIndex >= len(m.emails) {
					m.emailFocusIndex = len(m.emails) - 1
				}
				return m.updateEmailFocus()
			}
		case "tab", "down":
			m.emailFocusIndex++
			if m.emailFocusIndex >= len(m.emailInputs) {
				m.emailFocusIndex = 0
			}
			return m.updateEmailFocus()
		case "shift+tab", "up":
			m.emailFocusIndex--
			if m.emailFocusIndex < 0 {
				m.emailFocusIndex = len(m.emailInputs) - 1
			}
			return m.updateEmailFocus()
		}
	}
	
	// Update the focused email input
	if m.emailFocusIndex < len(m.emailInputs) {
		newInput, cmd := m.emailInputs[m.emailFocusIndex].Update(msg)
		m.emailInputs[m.emailFocusIndex] = newInput
		return m, cmd
	}
	
	return m, nil
}

func (m *ClientFormModel) updateMainFocus() (ClientFormModel, tea.Cmd) {
	cmds := []tea.Cmd{}
	
	// Update name input
	if m.focusIndex == 0 {
		cmd := m.nameInput.Focus()
		m.nameInput.PromptStyle = formInputStyle
		m.nameInput.TextStyle = formInputStyle
		cmds = append(cmds, cmd)
	} else {
		m.nameInput.Blur()
		m.nameInput.PromptStyle = dimStyle
		m.nameInput.TextStyle = dimStyle
	}
	
	// Update address input
	if m.focusIndex == 1 {
		cmd := m.addressInput.Focus()
		m.addressInput.PromptStyle = formInputStyle
		m.addressInput.TextStyle = formInputStyle
		cmds = append(cmds, cmd)
	} else {
		m.addressInput.Blur()
		m.addressInput.PromptStyle = dimStyle
		m.addressInput.TextStyle = dimStyle
	}
	
	// Update hourly rate input
	if m.focusIndex == 2 {
		cmd := m.hourlyRateInput.Focus()
		m.hourlyRateInput.PromptStyle = formInputStyle
		m.hourlyRateInput.TextStyle = formInputStyle
		cmds = append(cmds, cmd)
	} else {
		m.hourlyRateInput.Blur()
		m.hourlyRateInput.PromptStyle = dimStyle
		m.hourlyRateInput.TextStyle = dimStyle
	}
	
	return *m, tea.Batch(cmds...)
}

func (m *ClientFormModel) updateEmailFocus() (ClientFormModel, tea.Cmd) {
	cmds := []tea.Cmd{}
	
	for i := range m.emailInputs {
		if i == m.emailFocusIndex {
			cmd := m.emailInputs[i].Focus()
			m.emailInputs[i].PromptStyle = formInputStyle
			m.emailInputs[i].TextStyle = formInputStyle
			cmds = append(cmds, cmd)
		} else {
			m.emailInputs[i].Blur()
			m.emailInputs[i].PromptStyle = dimStyle
			m.emailInputs[i].TextStyle = dimStyle
		}
	}
	
	return *m, tea.Batch(cmds...)
}

func (m *ClientFormModel) updateMainInputs(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{}
	
	if m.focusIndex == 0 {
		newInput, cmd := m.nameInput.Update(msg)
		m.nameInput = newInput
		cmds = append(cmds, cmd)
	} else if m.focusIndex == 1 {
		newInput, cmd := m.addressInput.Update(msg)
		m.addressInput = newInput
		cmds = append(cmds, cmd)
	} else if m.focusIndex == 2 {
		newInput, cmd := m.hourlyRateInput.Update(msg)
		m.hourlyRateInput = newInput
		cmds = append(cmds, cmd)
	}
	
	return tea.Batch(cmds...)
}

func (m *ClientFormModel) saveClient() error {
	name := strings.TrimSpace(m.nameInput.Value())
	address := strings.TrimSpace(m.addressInput.Value())
	
	if name == "" {
		return fmt.Errorf("name is required")
	}
	
	// Collect non-empty emails
	validEmails := []string{}
	for i, input := range m.emailInputs {
		email := strings.TrimSpace(input.Value())
		if email == "" && i < len(m.emails) {
			email = strings.TrimSpace(m.emails[i])
		}
		if email != "" {
			validEmails = append(validEmails, email)
		}
	}
	
	if len(validEmails) == 0 {
		return fmt.Errorf("at least one email is required")
	}
	
	// Parse hourly rate
	hourlyRate := decimal.Zero
	if rateStr := strings.TrimSpace(m.hourlyRateInput.Value()); rateStr != "" {
		rate, err := decimal.NewFromString(rateStr)
		if err != nil {
			return fmt.Errorf("invalid hourly rate: %v", err)
		}
		hourlyRate = rate
	}
	
	if m.isEdit {
		m.client.Update(name, address, validEmails, hourlyRate)
		return m.storage.UpdateClient(m.client)
	}
	
	client := models.NewClient(name, address, validEmails, hourlyRate)
	return m.storage.SaveClient(client)
}

func (m ClientFormModel) View() string {
	if m.mode == clientFormModeManageEmails {
		return m.viewManageEmails()
	}
	
	return m.viewEditMode()
}

func (m ClientFormModel) viewEditMode() string {
	var s strings.Builder
	
	title := "New Client"
	if m.isEdit {
		title = "Edit Client"
	}
	s.WriteString(titleStyle.Render(title) + "\n\n")
	
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
	}
	
	// Name field
	s.WriteString(formLabelStyle.Render("Name:"))
	s.WriteString(m.nameInput.View() + "\n")
	
	// Address field
	s.WriteString(formLabelStyle.Render("Address:"))
	s.WriteString(m.addressInput.View() + "\n")
	
	// Hourly rate field
	s.WriteString(formLabelStyle.Render("Hourly Rate:"))
	s.WriteString(m.hourlyRateInput.View() + "\n")
	
	// Emails summary with manage button
	emailCount := 0
	for _, input := range m.emailInputs {
		if strings.TrimSpace(input.Value()) != "" {
			emailCount++
		}
	}
	emailsText := fmt.Sprintf("%d email(s)", emailCount)
	manageButton := "[ Manage Emails ]"
	if m.focusIndex == 3 {
		manageButton = selectedStyle.Render(manageButton)
	}
	s.WriteString(formLabelStyle.Render("Emails:") + emailsText + " " + manageButton + "\n\n")
	
	// Save button
	saveButton := "[ Save ]"
	if m.focusIndex == 4 {
		saveButton = selectedStyle.Render(saveButton)
	}
	s.WriteString(saveButton)
	
	s.WriteString("\n\n" + helpStyle.Render("tab/shift+tab navigate • enter select • esc cancel"))
	
	return appStyle.Render(s.String())
}

func (m ClientFormModel) viewManageEmails() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Manage Emails") + "\n\n")
	
	if len(m.emailInputs) == 0 {
		s.WriteString(dimStyle.Render("No emails. Press '+' to add.") + "\n")
	} else {
		for i, input := range m.emailInputs {
			if i == m.emailFocusIndex {
				s.WriteString("> ")
			} else {
				s.WriteString("  ")
			}
			s.WriteString(input.View() + "\n")
		}
	}
	
	s.WriteString("\n" + helpStyle.Render("+ add email • - remove email • tab/↑/↓ navigate • esc done"))
	
	return appStyle.Render(s.String())
}