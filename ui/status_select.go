package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/invoicer/models"
)

type StatusSelectedMsg struct {
	Status models.InvoiceStatus
	Reason string
}

type CancelStatusChangeMsg struct{}

type StatusSelectModel struct {
	currentStatus models.InvoiceStatus
	selectedStatus models.InvoiceStatus
	reasonInput   textinput.Model
	focusIndex    int
	statusOptions []models.InvoiceStatus
}

func NewStatusSelectModel(currentStatus models.InvoiceStatus) StatusSelectModel {
	ti := textinput.New()
	ti.Placeholder = "Optional reason for status change..."
	ti.CharLimit = 200
	ti.Width = 50
	
	return StatusSelectModel{
		currentStatus: currentStatus,
		selectedStatus: currentStatus,
		reasonInput:   ti,
		focusIndex:    0,
		statusOptions: []models.InvoiceStatus{
			models.StatusDraft,
			models.StatusSent,
			models.StatusPaid,
			models.StatusOverdue,
		},
	}
}

func (m StatusSelectModel) Init() tea.Cmd {
	return nil
}

func (m StatusSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return CancelStatusChangeMsg{} }
		case "enter":
			if m.focusIndex < len(m.statusOptions) {
				// Confirm status selection
				return m, func() tea.Msg {
					return StatusSelectedMsg{
						Status: m.selectedStatus,
						Reason: m.reasonInput.Value(),
					}
				}
			}
			// If focused on reason input, do nothing on enter
			return m, nil
		case "up", "k":
			if m.focusIndex > 0 {
				m.focusIndex--
				if m.focusIndex < len(m.statusOptions) {
					m.selectedStatus = m.statusOptions[m.focusIndex]
				}
			}
		case "down", "j":
			if m.focusIndex < len(m.statusOptions) {
				m.focusIndex++
			}
		case "tab", "shift+tab":
			// Toggle between status options and reason input
			if m.focusIndex < len(m.statusOptions) {
				m.focusIndex = len(m.statusOptions)
				m.reasonInput.Focus()
			} else {
				m.focusIndex = 0
				m.selectedStatus = m.statusOptions[0]
				m.reasonInput.Blur()
			}
		}
		
		// Handle text input updates when focused
		if m.focusIndex == len(m.statusOptions) {
			var cmd tea.Cmd
			m.reasonInput, cmd = m.reasonInput.Update(msg)
			return m, cmd
		}
	}
	
	return m, nil
}

func (m StatusSelectModel) View() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Change Invoice Status") + "\n\n")
	s.WriteString(formLabelStyle.Render("Current Status: ") + getStatusStyle(m.currentStatus).Render(string(m.currentStatus)) + "\n\n")
	s.WriteString(formLabelStyle.Render("Select New Status:") + "\n")
	
	// Status options
	for i, status := range m.statusOptions {
		cursor := "  "
		if i == m.focusIndex && m.focusIndex < len(m.statusOptions) {
			cursor = "> "
			if status == m.selectedStatus {
				s.WriteString(selectedStyle.Render(cursor + string(status)) + "\n")
			} else {
				s.WriteString(cursor + string(status) + "\n")
			}
		} else if status == m.selectedStatus {
			s.WriteString(cursor + selectedStyle.Render(string(status)) + "\n")
		} else {
			s.WriteString(cursor + string(status) + "\n")
		}
	}
	
	// Reason input
	s.WriteString("\n" + formLabelStyle.Render("Reason (optional):") + "\n")
	if m.focusIndex == len(m.statusOptions) {
		s.WriteString(m.reasonInput.View() + "\n")
	} else {
		s.WriteString(dimStyle.Render(m.reasonInput.View()) + "\n")
	}
	
	s.WriteString("\n" + helpStyle.Render("↑/↓ navigate • tab switch to reason • enter confirm • esc cancel"))
	
	return appStyle.Render(s.String())
}

func getStatusStyle(status models.InvoiceStatus) lipgloss.Style {
	switch status {
	case models.StatusDraft:
		return statusDraftStyle
	case models.StatusSent:
		return statusSentStyle
	case models.StatusPaid:
		return statusPaidStyle
	case models.StatusOverdue:
		return statusOverdueStyle
	default:
		return normalStyle
	}
}