package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/invoicer/models"
)

type menuChoice int

const (
	menuClients menuChoice = iota
	menuInvoices
	menuExit
)

var menuItems = []string{
	"Manage Clients",
	"Manage Invoices",
	"Exit",
}

type MainMenuModel struct {
	cursor  int
	choice  menuChoice
	storage models.Storage
}

func NewMainMenuModel(storage models.Storage) MainMenuModel {
	return MainMenuModel{
		storage: storage,
	}
}

func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "enter":
			m.choice = menuChoice(m.cursor)
			switch m.choice {
			case menuClients:
				return NewClientListModel(m.storage), nil
			case menuInvoices:
				return NewInvoiceListModel(m.storage), nil
			case menuExit:
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m MainMenuModel) View() string {
	s := titleStyle.Render("Invoice Manager") + "\n\n"
	
	for i, item := range menuItems {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			s += selectedListItemStyle.Render(cursor + item) + "\n"
		} else {
			s += listItemStyle.Render(item) + "\n"
		}
	}
	
	s += "\n" + helpStyle.Render("↑/k up • ↓/j down • enter select • q quit")
	
	return appStyle.Render(s)
}

type BackToMenuMsg struct{}

func BackToMenu() tea.Msg {
	return BackToMenuMsg{}
}