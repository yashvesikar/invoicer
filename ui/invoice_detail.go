package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/invoicer/models"
)

type InvoiceDetailModel struct {
	invoice *models.Invoice
	storage models.Storage
	width   int
}

func NewInvoiceDetailModel(storage models.Storage, invoice *models.Invoice) InvoiceDetailModel {
	return InvoiceDetailModel{
		invoice: invoice,
		storage: storage,
		width:   80,
	}
}

func (m InvoiceDetailModel) Init() tea.Cmd {
	return nil
}

func (m InvoiceDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewInvoiceListModel(m.storage), func() tea.Msg { return BackToInvoiceListMsg{} }
		case "e":
			return NewInvoiceFormModel(m.storage, m.invoice), nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width - 4
		if m.width > 100 {
			m.width = 100
		}
	}
	return m, nil
}

func (m InvoiceDetailModel) View() string {
	var s strings.Builder
	
	titleBlock := titleStyle.Render(fmt.Sprintf("Invoice %s", m.invoice.Number))
	s.WriteString(titleBlock + "\n\n")
	
	leftColStyle := lipgloss.NewStyle().Width(m.width / 2)
	rightColStyle := lipgloss.NewStyle().Width(m.width/2).Align(lipgloss.Right)
	
	leftCol := []string{
		formLabelStyle.Render("Client:") + " " + m.invoice.ClientName,
		formLabelStyle.Render("Date:") + " " + m.invoice.Date.Format("January 2, 2006"),
		formLabelStyle.Render("Due Date:") + " " + m.invoice.DueDate.Format("January 2, 2006"),
	}
	
	statusStyle := normalStyle
	switch m.invoice.Status {
	case models.StatusDraft:
		statusStyle = statusDraftStyle
	case models.StatusSent:
		statusStyle = statusSentStyle
	case models.StatusPaid:
		statusStyle = statusPaidStyle
	case models.StatusOverdue:
		statusStyle = statusOverdueStyle
	}
	
	rightCol := []string{
		formLabelStyle.Render("Status:") + " " + statusStyle.Render(string(m.invoice.Status)),
		"",
		"",
	}
	
	for i := 0; i < 3; i++ {
		row := leftColStyle.Render(leftCol[i]) + rightColStyle.Render(rightCol[i])
		s.WriteString(row + "\n")
	}
	
	s.WriteString("\n" + subtitleStyle.Render("Line Items") + "\n")
	s.WriteString(strings.Repeat("─", m.width) + "\n")
	
	if len(m.invoice.LineItems) == 0 {
		s.WriteString(dimStyle.Render("No line items") + "\n")
	} else {
		headers := []string{"Description", "Qty", "Unit Price", "Total"}
		widths := []int{m.width - 30, 8, 10, 12}
		
		headerRow := ""
		for i, h := range headers {
			style := tableHeaderStyle.Copy().Width(widths[i])
			if i > 0 {
				style = style.Align(lipgloss.Right)
			}
			headerRow += style.Render(h)
		}
		s.WriteString(headerRow + "\n")
		
		for _, item := range m.invoice.LineItems {
			cells := []string{
				truncate(item.Description, widths[0]-2),
				fmt.Sprintf("%.2f", item.Quantity.InexactFloat64()),
				fmt.Sprintf("$%.2f", item.UnitPrice.InexactFloat64()),
				fmt.Sprintf("$%.2f", item.Total.InexactFloat64()),
			}
			
			row := ""
			for i, cell := range cells {
				style := tableCellStyle.Copy().Width(widths[i])
				if i > 0 {
					style = style.Align(lipgloss.Right)
				}
				row += style.Render(cell)
			}
			s.WriteString(row + "\n")
		}
	}
	
	s.WriteString(strings.Repeat("─", m.width) + "\n")
	
	summaryStyle := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Right)
	s.WriteString(summaryStyle.Render(fmt.Sprintf("Subtotal: $%.2f", m.invoice.Subtotal.InexactFloat64())) + "\n")
	
	if m.invoice.DiscountRate.GreaterThan(models.DecimalZero) {
		s.WriteString(summaryStyle.Render(fmt.Sprintf("Discount (%.1f%%): -$%.2f", 
			m.invoice.DiscountRate.InexactFloat64(), 
			m.invoice.Discount.InexactFloat64())) + "\n")
	}
	
	if m.invoice.TaxRate.GreaterThan(models.DecimalZero) {
		s.WriteString(summaryStyle.Render(fmt.Sprintf("Tax (%.1f%%): $%.2f", 
			m.invoice.TaxRate.InexactFloat64(), 
			m.invoice.Tax.InexactFloat64())) + "\n")
	}
	
	totalStyle := summaryStyle.Copy().Bold(true)
	s.WriteString(totalStyle.Render(fmt.Sprintf("Total: $%.2f", m.invoice.Total.InexactFloat64())) + "\n")
	
	s.WriteString("\n" + helpStyle.Render("e edit • esc back • q quit"))
	
	return appStyle.Render(s.String())
}