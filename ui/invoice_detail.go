package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/invoicer/config"
	"github.com/user/invoicer/export"
	"github.com/user/invoicer/models"
)

type invoiceDetailMode int

const (
	invoiceDetailModeView invoiceDetailMode = iota
	invoiceDetailModeExportLocation
)

type InvoiceDetailModel struct {
	invoice         *models.Invoice
	storage         models.Storage
	config          *config.Config
	width           int
	message         string
	isError         bool
	mode            invoiceDetailMode
	exportLocationModel ExportLocationModel
}

func NewInvoiceDetailModel(storage models.Storage, cfg *config.Config, invoice *models.Invoice) InvoiceDetailModel {
	return InvoiceDetailModel{
		invoice: invoice,
		storage: storage,
		config:  cfg,
		width:   80,
		mode:    invoiceDetailModeView,
	}
}

func (m InvoiceDetailModel) Init() tea.Cmd {
	return nil
}

func (m InvoiceDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case invoiceDetailModeView:
		return m.updateView(msg)
	case invoiceDetailModeExportLocation:
		return m.updateExportLocation(msg)
	}
	return m, nil
}

func (m InvoiceDetailModel) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewInvoiceListModel(m.storage, m.config), func() tea.Msg { return BackToInvoiceListMsg{} }
		case "e":
			return NewInvoiceFormModel(m.storage, m.config, m.invoice), nil
		case "p":
			// Switch to export location mode
			m.mode = invoiceDetailModeExportLocation
			m.exportLocationModel = NewExportLocationModel(m.invoice.Number)
			return m, m.exportLocationModel.Init()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width - 4
		if m.width > 100 {
			m.width = 100
		}
	}
	
	return m, nil
}

func (m InvoiceDetailModel) updateExportLocation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ExportLocationSelectedMsg:
		// Export to selected location
		client, err := m.storage.GetClient(m.invoice.ClientID)
		if err != nil {
			m.message = fmt.Sprintf("Error loading client: %v", err)
			m.isError = true
			m.mode = invoiceDetailModeView
			return m, nil
		}
		
		// Get template path
		templatePath := filepath.Join(m.config.TemplatesDir(), "invoice.tex")
		
		err = export.ExportInvoiceToPDF(m.invoice, client, m.config, msg.Path, templatePath)
		if err != nil {
			m.message = fmt.Sprintf("Error exporting PDF: %v", err)
			m.isError = true
		} else {
			exportPath := export.GetExportPath(m.invoice, msg.Path)
			m.message = fmt.Sprintf("Invoice exported to: %s", exportPath)
			m.isError = false
		}
		m.mode = invoiceDetailModeView
		return m, nil
		
	case CancelExportMsg:
		// Cancel export and go back to view mode
		m.mode = invoiceDetailModeView
		return m, nil
		
	default:
		// Pass through to export location model
		var cmd tea.Cmd
		model, cmd := m.exportLocationModel.Update(msg)
		m.exportLocationModel = model.(ExportLocationModel)
		return m, cmd
	}
}

func (m InvoiceDetailModel) View() string {
	switch m.mode {
	case invoiceDetailModeView:
		return m.viewDetail()
	case invoiceDetailModeExportLocation:
		return m.exportLocationModel.View()
	}
	return ""
}

func (m InvoiceDetailModel) viewDetail() string {
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
	
	if m.message != "" {
		s.WriteString("\n")
		if m.isError {
			s.WriteString(errorStyle.Render(m.message) + "\n")
		} else {
			s.WriteString(successStyle.Render(m.message) + "\n")
		}
	}
	
	s.WriteString("\n" + helpStyle.Render("e edit • p export PDF • esc back • q quit"))
	
	return appStyle.Render(s.String())
}