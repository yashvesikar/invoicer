package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/invoicer/models"
)

type clientListMode int

const (
	clientListModeView clientListMode = iota
	clientListModeConfirmDelete
)

type ClientListModel struct {
	clients         []models.Client
	cursor          int
	storage         models.Storage
	mode            clientListMode
	selectedForDelete string
	err             error
}

func NewClientListModel(storage models.Storage) ClientListModel {
	m := ClientListModel{
		storage: storage,
		mode:    clientListModeView,
	}
	m.loadClients()
	return m
}

func (m *ClientListModel) loadClients() {
	clients, err := m.storage.GetAllClients()
	if err != nil {
		m.err = err
		return
	}
	m.clients = clients
	m.err = nil
}

func (m ClientListModel) Init() tea.Cmd {
	return nil
}

func (m ClientListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case clientListModeView:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				return NewMainMenuModel(m.storage), nil
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.clients)-1 {
					m.cursor++
				}
			case "a":
				return NewClientFormModel(m.storage, nil), nil
			case "e":
				if len(m.clients) > 0 {
					return NewClientFormModel(m.storage, &m.clients[m.cursor]), nil
				}
			case "d":
				if len(m.clients) > 0 {
					m.mode = clientListModeConfirmDelete
					m.selectedForDelete = m.clients[m.cursor].ID
				}
			case "enter":
				if len(m.clients) > 0 {
					return NewClientFormModel(m.storage, &m.clients[m.cursor]), nil
				}
			}
		case clientListModeConfirmDelete:
			switch msg.String() {
			case "y":
				err := m.storage.DeleteClient(m.selectedForDelete)
				if err != nil {
					m.err = err
				}
				m.loadClients()
				m.mode = clientListModeView
				if m.cursor >= len(m.clients) && m.cursor > 0 {
					m.cursor = len(m.clients) - 1
				}
			case "n", "esc":
				m.mode = clientListModeView
			}
		}
	case BackToClientListMsg:
		m.loadClients()
	}
	return m, nil
}

func (m ClientListModel) View() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Clients") + "\n\n")
	
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
	}
	
	if m.mode == clientListModeConfirmDelete {
		client := m.getClientByID(m.selectedForDelete)
		if client != nil {
			s.WriteString(errorStyle.Render(fmt.Sprintf("Delete client '%s'? (y/n)", client.Name)) + "\n")
		}
		return appStyle.Render(s.String())
	}
	
	if len(m.clients) == 0 {
		s.WriteString(dimStyle.Render("No clients found. Press 'a' to add a new client.") + "\n")
	} else {
		headers := []string{"Name", "Email", "Address"}
		widths := []int{25, 30, 40}
		
		headerRow := ""
		for i, h := range headers {
			headerRow += tableCellStyle.Width(widths[i]).Render(h)
		}
		s.WriteString(tableHeaderStyle.Render(headerRow) + "\n")
		
		for i, client := range m.clients {
			row := ""
			email := ""
			if len(client.Emails) > 0 {
				email = client.Emails[0]
			}
			cells := []string{
				truncate(client.Name, widths[0]-2),
				truncate(email, widths[1]-2),
				truncate(client.Address, widths[2]-2),
			}
			
			for j, cell := range cells {
				style := tableCellStyle.Width(widths[j])
				if i == m.cursor {
					style = style.Inherit(selectedStyle)
				}
				row += style.Render(cell)
			}
			
			if i == m.cursor {
				s.WriteString("> " + row + "\n")
			} else {
				s.WriteString("  " + row + "\n")
			}
		}
	}
	
	s.WriteString("\n" + helpStyle.Render("a add • e edit • d delete • ↑/k up • ↓/j down • esc back • q quit"))
	
	return appStyle.Render(s.String())
}

func (m ClientListModel) getClientByID(id string) *models.Client {
	for _, c := range m.clients {
		if c.ID == id {
			return &c
		}
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

type BackToClientListMsg struct{}