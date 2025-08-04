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
	invoiceDetailModeStatusSelect
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
	statusSelectModel   StatusSelectModel
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
	case invoiceDetailModeStatusSelect:
		return m.updateStatusSelect(msg)
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
		case "s":
			// Switch to status select mode
			m.mode = invoiceDetailModeStatusSelect
			m.statusSelectModel = NewStatusSelectModel(m.invoice.Status)
			return m, m.statusSelectModel.Init()
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

func (m InvoiceDetailModel) updateStatusSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusSelectedMsg:
		// Update invoice status
		oldStatus := m.invoice.Status
		err := m.invoice.UpdateStatus(msg.Status, msg.Reason)
		if err != nil {
			m.message = fmt.Sprintf("Error updating status: %v", err)
			m.isError = true
			m.mode = invoiceDetailModeView
			return m, nil
		}
		
		// Save the updated invoice
		err = m.storage.UpdateInvoice(m.invoice)
		if err != nil {
			m.message = fmt.Sprintf("Error saving invoice: %v", err)
			m.isError = true
			m.mode = invoiceDetailModeView
			return m, nil
		}
		
		// Log the status change in audit
		entry := models.NewAuditEntry(
			m.invoice.ID,
			m.invoice.Number,
			oldStatus,
			m.invoice.Status,
			msg.Reason,
		)
		
		err = m.storage.SaveAuditEntry(entry)
		if err != nil {
			// Log error but don't fail the operation
			m.message = fmt.Sprintf("Status updated but audit log failed: %v", err)
			m.isError = true
		} else {
			m.message = fmt.Sprintf("Status changed from %s to %s", oldStatus, msg.Status)
			m.isError = false
		}
		
		m.mode = invoiceDetailModeView
		return m, nil
		
	case CancelStatusChangeMsg:
		// Cancel status change and go back to view mode
		m.mode = invoiceDetailModeView
		return m, nil
		
	default:
		// Pass through to status select model
		var cmd tea.Cmd
		model, cmd := m.statusSelectModel.Update(msg)
		m.statusSelectModel = model.(StatusSelectModel)
		return m, cmd
	}
}

func (m InvoiceDetailModel) View() string {
	switch m.mode {
	case invoiceDetailModeView:
		return m.viewDetail()
	case invoiceDetailModeExportLocation:
		return m.exportLocationModel.View()
	case invoiceDetailModeStatusSelect:
		return m.statusSelectModel.View()
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
		// Calculate dynamic column widths
		descWidth := m.width - 35 // Reserve space for numeric columns
		if descWidth < 30 {
			descWidth = 30
		}
		
		headers := []string{"Description", "Qty", "Unit Price", "Total"}
		widths := []int{descWidth, 10, 12, 13}
		
		// Header row with better spacing
		headerRow := ""
		for i, h := range headers {
			style := tableHeaderStyle.Copy().Width(widths[i]).Padding(0, 1)
			if i > 0 {
				style = style.Align(lipgloss.Right)
			}
			headerRow += style.Render(h)
		}
		s.WriteString(headerRow + "\n")
		
		// Add a subtle separator after headers
		s.WriteString(dimStyle.Render(strings.Repeat("─", m.width)) + "\n")
		
		// Render line items with alternating background
		for idx, item := range m.invoice.LineItems {
			cells := []string{
				truncate(item.Description, widths[0]-4),
				fmt.Sprintf("%.2f", item.Quantity.InexactFloat64()),
				fmt.Sprintf("$%.2f", item.UnitPrice.InexactFloat64()),
				fmt.Sprintf("$%.2f", item.Total.InexactFloat64()),
			}
			
			row := ""
			for i, cell := range cells {
				style := tableCellStyle.Copy().Width(widths[i]).Padding(0, 1)
				if i > 0 {
					style = style.Align(lipgloss.Right)
				}
				// Add subtle alternating background
				if idx%2 == 1 {
					style = style.Background(lipgloss.Color("236"))
				}
				row += style.Render(cell)
			}
			s.WriteString(row + "\n")
		}
	}
	
	s.WriteString(strings.Repeat("─", m.width) + "\n")
	
	// Create a summary section with better formatting
	summaryWidth := 30
	summaryOffset := m.width - summaryWidth
	
	// Helper function to format summary lines
	formatSummaryLine := func(label, amount string, isBold bool) string {
		labelStyle := lipgloss.NewStyle().Width(15).Align(lipgloss.Right)
		amountStyle := lipgloss.NewStyle().Width(15).Align(lipgloss.Right)
		if isBold {
			labelStyle = labelStyle.Bold(true)
			amountStyle = amountStyle.Bold(true)
		}
		return strings.Repeat(" ", summaryOffset) + 
			labelStyle.Render(label) + 
			amountStyle.Render(amount)
	}
	
	s.WriteString(formatSummaryLine("Subtotal:", fmt.Sprintf("$%.2f", m.invoice.Subtotal.InexactFloat64()), false) + "\n")
	
	if m.invoice.DiscountRate.GreaterThan(models.DecimalZero) {
		s.WriteString(formatSummaryLine(
			fmt.Sprintf("Discount (%.1f%%):", m.invoice.DiscountRate.InexactFloat64()),
			fmt.Sprintf("-$%.2f", m.invoice.Discount.InexactFloat64()),
			false,
		) + "\n")
	}
	
	if m.invoice.TaxRate.GreaterThan(models.DecimalZero) {
		s.WriteString(formatSummaryLine(
			fmt.Sprintf("Tax (%.1f%%):", m.invoice.TaxRate.InexactFloat64()),
			fmt.Sprintf("$%.2f", m.invoice.Tax.InexactFloat64()),
			false,
		) + "\n")
	}
	
	// Add a separator before total
	s.WriteString(strings.Repeat(" ", summaryOffset) + dimStyle.Render(strings.Repeat("─", summaryWidth)) + "\n")
	
	s.WriteString(formatSummaryLine("Total:", fmt.Sprintf("$%.2f", m.invoice.Total.InexactFloat64()), true) + "\n")
	
	if m.message != "" {
		s.WriteString("\n")
		if m.isError {
			s.WriteString(errorStyle.Render(m.message) + "\n")
		} else {
			s.WriteString(successStyle.Render(m.message) + "\n")
		}
	}
	
	// Show audit history
	auditEntries, err := m.storage.GetAuditEntries(m.invoice.ID)
	if err == nil && len(auditEntries) > 0 {
		s.WriteString("\n\n" + subtitleStyle.Render("Status History") + "\n")
		s.WriteString(strings.Repeat("─", m.width) + "\n")
		
		for i, entry := range auditEntries {
			if i > 0 {
				s.WriteString(dimStyle.Render(strings.Repeat("·", m.width)) + "\n")
			}
			
			timeStr := entry.ChangedAt.Format("Jan 2, 2006 3:04 PM")
			s.WriteString(formLabelStyle.Render("Date:") + " " + timeStr + "\n")
			s.WriteString(formLabelStyle.Render("Changed:") + " " + 
				getStatusStyle(entry.OldStatus).Render(string(entry.OldStatus)) + 
				" → " + 
				getStatusStyle(entry.NewStatus).Render(string(entry.NewStatus)) + "\n")
			
			if entry.Reason != "" {
				s.WriteString(formLabelStyle.Render("Reason:") + " " + entry.Reason + "\n")
			}
		}
	}
	
	s.WriteString("\n" + helpStyle.Render("e edit • s change status • p export PDF • esc back • q quit"))
	
	return appStyle.Render(s.String())
}