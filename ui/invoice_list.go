package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/invoicer/config"
	"github.com/user/invoicer/models"
)

type invoiceListMode int

const (
	invoiceListModeView invoiceListMode = iota
	invoiceListModeConfirmDelete
)

type InvoiceListModel struct {
	invoices          []models.Invoice
	cursor            int
	storage           models.Storage
	config            *config.Config
	mode              invoiceListMode
	selectedForDelete string
	err               error
}

func NewInvoiceListModel(storage models.Storage, cfg *config.Config) InvoiceListModel {
	m := InvoiceListModel{
		storage: storage,
		config:  cfg,
		mode:    invoiceListModeView,
	}
	m.loadInvoices()
	return m
}

func (m *InvoiceListModel) loadInvoices() {
	invoices, err := m.storage.GetAllInvoices()
	if err != nil {
		m.err = err
		return
	}
	m.invoices = invoices
	m.err = nil
}

func (m InvoiceListModel) Init() tea.Cmd {
	return nil
}

func (m InvoiceListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case invoiceListModeView:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				return NewMainMenuModel(m.storage, m.config), nil
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.invoices)-1 {
					m.cursor++
				}
			case "a":
				clients, err := m.storage.GetAllClients()
				if err != nil {
					m.err = err
					return m, nil
				}
				if len(clients) == 0 {
					m.err = fmt.Errorf("no clients found. Please add a client first")
					return m, nil
				}
				return NewInvoiceFormModel(m.storage, m.config, nil), nil
			case "e":
				if len(m.invoices) > 0 {
					return NewInvoiceFormModel(m.storage, m.config, &m.invoices[m.cursor]), nil
				}
			case "v":
				if len(m.invoices) > 0 {
					return NewInvoiceDetailModel(m.storage, m.config, &m.invoices[m.cursor]), nil
				}
			case "d":
				if len(m.invoices) > 0 {
					m.mode = invoiceListModeConfirmDelete
					m.selectedForDelete = m.invoices[m.cursor].ID
				}
			case "enter":
				if len(m.invoices) > 0 {
					return NewInvoiceDetailModel(m.storage, m.config, &m.invoices[m.cursor]), nil
				}
			}
		case invoiceListModeConfirmDelete:
			switch msg.String() {
			case "y":
				err := m.storage.DeleteInvoice(m.selectedForDelete)
				if err != nil {
					m.err = err
				}
				m.loadInvoices()
				m.mode = invoiceListModeView
				if m.cursor >= len(m.invoices) && m.cursor > 0 {
					m.cursor = len(m.invoices) - 1
				}
			case "n", "esc":
				m.mode = invoiceListModeView
			}
		}
	case BackToInvoiceListMsg:
		m.loadInvoices()
	}
	return m, nil
}

func (m InvoiceListModel) View() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Invoices") + "\n\n")
	
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
		m.err = nil
	}
	
	if m.mode == invoiceListModeConfirmDelete {
		invoice := m.getInvoiceByID(m.selectedForDelete)
		if invoice != nil {
			s.WriteString(errorStyle.Render(fmt.Sprintf("Delete invoice '%s'? (y/n)", invoice.Number)) + "\n")
		}
		return appStyle.Render(s.String())
	}
	
	if len(m.invoices) == 0 {
		s.WriteString(dimStyle.Render("No invoices found. Press 'a' to create a new invoice.") + "\n")
	} else {
		headers := []string{"Number", "Client", "Date", "Total", "Status"}
		widths := []int{10, 25, 12, 12, 10}
		
		headerRow := ""
		for i, h := range headers {
			headerRow += tableCellStyle.Width(widths[i]).Render(h)
		}
		s.WriteString(tableHeaderStyle.Render(headerRow) + "\n")
		
		for i, invoice := range m.invoices {
			row := ""
			cells := []string{
				invoice.Number,
				truncate(invoice.ClientName, widths[1]-2),
				invoice.Date.Format("2006-01-02"),
				fmt.Sprintf("$%.2f", invoice.Total.InexactFloat64()),
				string(invoice.Status),
			}
			
			for j, cell := range cells {
				style := tableCellStyle.Width(widths[j])
				if i == m.cursor {
					style = style.Inherit(selectedStyle)
				}
				
				if j == 4 {
					switch invoice.Status {
					case models.StatusDraft:
						style = style.Inherit(statusDraftStyle)
					case models.StatusSent:
						style = style.Inherit(statusSentStyle)
					case models.StatusPaid:
						style = style.Inherit(statusPaidStyle)
					case models.StatusOverdue:
						style = style.Inherit(statusOverdueStyle)
					}
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
	
	s.WriteString("\n" + helpStyle.Render("a add • e edit • v view • d delete • ↑/k up • ↓/j down • esc back • q quit"))
	
	return appStyle.Render(s.String())
}

func (m InvoiceListModel) getInvoiceByID(id string) *models.Invoice {
	for _, inv := range m.invoices {
		if inv.ID == id {
			return &inv
		}
	}
	return nil
}

type BackToInvoiceListMsg struct{}