package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/shopspring/decimal"
	"github.com/user/invoicer/config"
	"github.com/user/invoicer/models"
)

type invoiceFormMode int

const (
	invoiceFormModeEditBasic invoiceFormMode = iota
	invoiceFormModeSelectClient
	invoiceFormModeEditLineItems
	invoiceFormModeAddLineItem
)

type InvoiceFormModel struct {
	storage            models.Storage
	config             *config.Config
	invoice            *models.Invoice
	isEdit             bool
	mode               invoiceFormMode
	
	clients            []models.Client
	clientCursor       int
	
	discountInput      textinput.Model
	taxInput           textinput.Model
	dueDaysInput       textinput.Model
	basicInputs        []textinput.Model
	basicFocusIndex    int
	
	lineItemInputs     []textinput.Model
	lineItemFocusIndex int
	lineItemCursor     int
	
	err                error
}

func NewInvoiceFormModel(storage models.Storage, cfg *config.Config, invoice *models.Invoice) InvoiceFormModel {
	discountInput := textinput.New()
	discountInput.Placeholder = "0"
	discountInput.Width = 10
	
	taxInput := textinput.New()
	taxInput.Placeholder = "0"
	taxInput.Width = 10
	
	dueDaysInput := textinput.New()
	dueDaysInput.Placeholder = "30"
	dueDaysInput.Width = 10
	
	m := InvoiceFormModel{
		storage:       storage,
		config:        cfg,
		invoice:       invoice,
		isEdit:        invoice != nil,
		mode:          invoiceFormModeEditBasic,
		discountInput: discountInput,
		taxInput:      taxInput,
		dueDaysInput:  dueDaysInput,
		basicInputs:   []textinput.Model{discountInput, taxInput, dueDaysInput},
	}
	
	if m.isEdit {
		m.discountInput.SetValue(fmt.Sprintf("%.1f", invoice.DiscountRate.InexactFloat64()))
		m.taxInput.SetValue(fmt.Sprintf("%.1f", invoice.TaxRate.InexactFloat64()))
		
		dueDays := int(invoice.DueDate.Sub(invoice.Date).Hours() / 24)
		m.dueDaysInput.SetValue(fmt.Sprintf("%d", dueDays))
	} else {
		clients, _ := storage.GetAllClients()
		m.clients = clients
		
		year := time.Now().Year()
		seq, _ := storage.GetNextInvoiceNumber(year)
		number := models.GenerateInvoiceNumber(year, seq)
		
		if len(clients) > 0 {
			m.invoice = models.NewInvoice(clients[0].ID, clients[0].Name, number)
		}
	}
	
	m.setupLineItemInputs()
	return m
}

func (m *InvoiceFormModel) setupLineItemInputs() {
	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.Width = 40
	descInput.Focus()
	
	qtyInput := textinput.New()
	qtyInput.Placeholder = "1"
	qtyInput.Width = 10
	
	priceInput := textinput.New()
	priceInput.Placeholder = "0.00"
	priceInput.Width = 10
	
	m.lineItemInputs = []textinput.Model{descInput, qtyInput, priceInput}
}

func (m InvoiceFormModel) Init() tea.Cmd {
	if m.mode == invoiceFormModeEditBasic {
		return textinput.Blink
	}
	return nil
}

func (m InvoiceFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case invoiceFormModeEditBasic:
			return m.updateBasicMode(msg)
		case invoiceFormModeSelectClient:
			return m.updateSelectClientMode(msg)
		case invoiceFormModeEditLineItems:
			return m.updateLineItemsMode(msg)
		case invoiceFormModeAddLineItem:
			return m.updateAddLineItemMode(msg)
		}
	}
	
	return m, nil
}

func (m InvoiceFormModel) updateBasicMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return NewInvoiceListModel(m.storage, m.config), func() tea.Msg { return BackToInvoiceListMsg{} }
	case "tab", "shift+tab", "enter", "up", "down":
		s := msg.String()
		
		totalFocusItems := len(m.basicInputs) + 1
		if !m.isEdit {
			totalFocusItems++
		}
		
		if s == "enter" && m.basicFocusIndex == totalFocusItems-1 {
			m.mode = invoiceFormModeEditLineItems
			return m, nil
		}
		
		if s == "enter" && !m.isEdit && m.basicFocusIndex == 0 {
			m.mode = invoiceFormModeSelectClient
			return m, nil
		}
		
		if s == "up" || s == "shift+tab" {
			m.basicFocusIndex--
		} else {
			m.basicFocusIndex++
		}
		
		if m.basicFocusIndex >= totalFocusItems {
			m.basicFocusIndex = 0
		} else if m.basicFocusIndex < 0 {
			m.basicFocusIndex = totalFocusItems - 1
		}
		
		return m.updateBasicFocus()
	}
	
	cmd := m.updateBasicInputs(msg)
	return m, cmd
}

func (m InvoiceFormModel) updateSelectClientMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = invoiceFormModeEditBasic
		return m, nil
	case "up", "k":
		if m.clientCursor > 0 {
			m.clientCursor--
		}
	case "down", "j":
		if m.clientCursor < len(m.clients)-1 {
			m.clientCursor++
		}
	case "enter":
		if len(m.clients) > 0 {
			client := m.clients[m.clientCursor]
			m.invoice.ClientID = client.ID
			m.invoice.ClientName = client.Name
		}
		m.mode = invoiceFormModeEditBasic
		return m, nil
	}
	return m, nil
}

func (m InvoiceFormModel) updateLineItemsMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = invoiceFormModeEditBasic
		return m, nil
	case "a":
		m.mode = invoiceFormModeAddLineItem
		m.setupLineItemInputs()
		return m, textinput.Blink
	case "d":
		if len(m.invoice.LineItems) > 0 && m.lineItemCursor < len(m.invoice.LineItems) {
			m.invoice.RemoveLineItem(m.invoice.LineItems[m.lineItemCursor].ID)
			if m.lineItemCursor >= len(m.invoice.LineItems) && m.lineItemCursor > 0 {
				m.lineItemCursor--
			}
		}
	case "up", "k":
		if m.lineItemCursor > 0 {
			m.lineItemCursor--
		}
	case "down", "j":
		if m.lineItemCursor < len(m.invoice.LineItems)-1 {
			m.lineItemCursor++
		}
	case "s":
		return m.saveInvoice()
	}
	return m, nil
}

func (m InvoiceFormModel) updateAddLineItemMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = invoiceFormModeEditLineItems
		return m, nil
	case "tab", "shift+tab", "enter":
		s := msg.String()
		
		if s == "enter" && m.lineItemFocusIndex == len(m.lineItemInputs) {
			if err := m.addLineItem(); err != nil {
				m.err = err
				return m, nil
			}
			m.mode = invoiceFormModeEditLineItems
			return m, nil
		}
		
		if s == "shift+tab" {
			m.lineItemFocusIndex--
		} else {
			m.lineItemFocusIndex++
		}
		
		if m.lineItemFocusIndex > len(m.lineItemInputs) {
			m.lineItemFocusIndex = 0
		} else if m.lineItemFocusIndex < 0 {
			m.lineItemFocusIndex = len(m.lineItemInputs)
		}
		
		return m.updateLineItemFocus()
	}
	
	cmd := m.updateLineItemInputs(msg)
	return m, cmd
}

func (m *InvoiceFormModel) updateBasicFocus() (InvoiceFormModel, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.basicInputs))
	
	offset := 0
	if !m.isEdit {
		offset = 1
	}
	
	for i := range m.basicInputs {
		if i == m.basicFocusIndex-offset && m.basicFocusIndex >= offset {
			cmds[i] = m.basicInputs[i].Focus()
			m.basicInputs[i].PromptStyle = formInputStyle
			m.basicInputs[i].TextStyle = formInputStyle
		} else {
			m.basicInputs[i].Blur()
			m.basicInputs[i].PromptStyle = dimStyle
			m.basicInputs[i].TextStyle = dimStyle
		}
	}
	
	return *m, tea.Batch(cmds...)
}

func (m *InvoiceFormModel) updateLineItemFocus() (InvoiceFormModel, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.lineItemInputs))
	
	for i := range m.lineItemInputs {
		if i == m.lineItemFocusIndex {
			cmds[i] = m.lineItemInputs[i].Focus()
			m.lineItemInputs[i].PromptStyle = formInputStyle
			m.lineItemInputs[i].TextStyle = formInputStyle
		} else {
			m.lineItemInputs[i].Blur()
			m.lineItemInputs[i].PromptStyle = dimStyle
			m.lineItemInputs[i].TextStyle = dimStyle
		}
	}
	
	return *m, tea.Batch(cmds...)
}

func (m *InvoiceFormModel) updateBasicInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.basicInputs))
	
	for i := range m.basicInputs {
		m.basicInputs[i], cmds[i] = m.basicInputs[i].Update(msg)
	}
	
	m.discountInput = m.basicInputs[0]
	m.taxInput = m.basicInputs[1]
	m.dueDaysInput = m.basicInputs[2]
	
	m.updateInvoiceRates()
	
	return tea.Batch(cmds...)
}

func (m *InvoiceFormModel) updateLineItemInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.lineItemInputs))
	
	for i := range m.lineItemInputs {
		m.lineItemInputs[i], cmds[i] = m.lineItemInputs[i].Update(msg)
	}
	
	return tea.Batch(cmds...)
}

func (m *InvoiceFormModel) updateInvoiceRates() {
	if discount, err := strconv.ParseFloat(m.discountInput.Value(), 64); err == nil {
		m.invoice.SetDiscountRate(decimal.NewFromFloat(discount))
	}
	
	if tax, err := strconv.ParseFloat(m.taxInput.Value(), 64); err == nil {
		m.invoice.SetTaxRate(decimal.NewFromFloat(tax))
	}
	
	if dueDays, err := strconv.Atoi(m.dueDaysInput.Value()); err == nil {
		m.invoice.DueDate = m.invoice.Date.AddDate(0, 0, dueDays)
	}
}

func (m *InvoiceFormModel) addLineItem() error {
	desc := strings.TrimSpace(m.lineItemInputs[0].Value())
	if desc == "" {
		return fmt.Errorf("description is required")
	}
	
	qtyStr := m.lineItemInputs[1].Value()
	if qtyStr == "" {
		qtyStr = "1"
	}
	qty, err := decimal.NewFromString(qtyStr)
	if err != nil {
		return fmt.Errorf("invalid quantity")
	}
	
	priceStr := m.lineItemInputs[2].Value()
	if priceStr == "" {
		priceStr = "0"
	}
	price, err := decimal.NewFromString(priceStr)
	if err != nil {
		return fmt.Errorf("invalid price")
	}
	
	item := models.NewLineItem(desc, qty, price)
	m.invoice.AddLineItem(*item)
	
	for i := range m.lineItemInputs {
		m.lineItemInputs[i].SetValue("")
	}
	
	return nil
}

func (m *InvoiceFormModel) saveInvoice() (tea.Model, tea.Cmd) {
	m.updateInvoiceRates()
	
	var err error
	if m.isEdit {
		err = m.storage.UpdateInvoice(m.invoice)
	} else {
		err = m.storage.SaveInvoice(m.invoice)
	}
	
	if err != nil {
		m.err = err
		return m, nil
	}
	
	return NewInvoiceListModel(m.storage, m.config), func() tea.Msg { return BackToInvoiceListMsg{} }
}

func (m InvoiceFormModel) View() string {
	switch m.mode {
	case invoiceFormModeSelectClient:
		return m.viewSelectClient()
	case invoiceFormModeEditLineItems:
		return m.viewLineItems()
	case invoiceFormModeAddLineItem:
		return m.viewAddLineItem()
	default:
		return m.viewBasic()
	}
}

func (m InvoiceFormModel) viewBasic() string {
	var s strings.Builder
	
	title := "New Invoice"
	if m.isEdit {
		title = fmt.Sprintf("Edit Invoice %s", m.invoice.Number)
	}
	s.WriteString(titleStyle.Render(title) + "\n\n")
	
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
	}
	
	if !m.isEdit {
		clientText := "[ Select Client ]"
		if m.basicFocusIndex == 0 {
			clientText = selectedStyle.Render(clientText)
		}
		s.WriteString(formLabelStyle.Render("Client:") + " " + m.invoice.ClientName + " " + clientText + "\n")
	} else {
		s.WriteString(formLabelStyle.Render("Client:") + " " + m.invoice.ClientName + "\n")
	}
	
	s.WriteString(formLabelStyle.Render("Invoice #:") + " " + m.invoice.Number + "\n")
	s.WriteString(formLabelStyle.Render("Date:") + " " + m.invoice.Date.Format("2006-01-02") + "\n\n")
	
	offset := 0
	if !m.isEdit {
		offset = 1
	}
	
	s.WriteString(formLabelStyle.Render("Discount %:"))
	s.WriteString(m.discountInput.View() + "\n")
	
	s.WriteString(formLabelStyle.Render("Tax %:"))
	s.WriteString(m.taxInput.View() + "\n")
	
	s.WriteString(formLabelStyle.Render("Due Days:"))
	s.WriteString(m.dueDaysInput.View() + "\n\n")
	
	nextButton := "[ Next: Line Items ]"
	if m.basicFocusIndex == len(m.basicInputs)+offset {
		nextButton = selectedStyle.Render(nextButton)
	}
	s.WriteString(nextButton)
	
	s.WriteString("\n\n" + helpStyle.Render("tab navigate • enter select • esc cancel"))
	
	return appStyle.Render(s.String())
}

func (m InvoiceFormModel) viewSelectClient() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Select Client") + "\n\n")
	
	if len(m.clients) == 0 {
		s.WriteString(dimStyle.Render("No clients found. Please add a client first.") + "\n")
	} else {
		for i, client := range m.clients {
			cursor := "  "
			if i == m.clientCursor {
				cursor = "> "
				s.WriteString(selectedListItemStyle.Render(cursor + client.Name) + "\n")
			} else {
				s.WriteString(listItemStyle.Render(client.Name) + "\n")
			}
		}
	}
	
	s.WriteString("\n" + helpStyle.Render("↑/k up • ↓/j down • enter select • esc cancel"))
	
	return appStyle.Render(s.String())
}

func (m InvoiceFormModel) viewLineItems() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render(fmt.Sprintf("Invoice %s - Line Items", m.invoice.Number)) + "\n\n")
	
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
	}
	
	if len(m.invoice.LineItems) == 0 {
		s.WriteString(dimStyle.Render("No line items. Press 'a' to add.") + "\n")
	} else {
		headers := []string{"Description", "Qty", "Price", "Total"}
		widths := []int{40, 10, 12, 12}
		
		headerRow := ""
		for i, h := range headers {
			headerRow += tableCellStyle.Width(widths[i]).Render(h)
		}
		s.WriteString(tableHeaderStyle.Render(headerRow) + "\n")
		
		for i, item := range m.invoice.LineItems {
			row := ""
			cells := []string{
				truncate(item.Description, widths[0]-2),
				fmt.Sprintf("%.2f", item.Quantity.InexactFloat64()),
				fmt.Sprintf("$%.2f", item.UnitPrice.InexactFloat64()),
				fmt.Sprintf("$%.2f", item.Total.InexactFloat64()),
			}
			
			for j, cell := range cells {
				style := tableCellStyle.Width(widths[j])
				if i == m.lineItemCursor {
					style = style.Inherit(selectedStyle)
				}
				row += style.Render(cell)
			}
			
			if i == m.lineItemCursor {
				s.WriteString("> " + row + "\n")
			} else {
				s.WriteString("  " + row + "\n")
			}
		}
		s.WriteString(strings.Repeat("─", 74) + "\n")
		s.WriteString(fmt.Sprintf("%*s", 74, fmt.Sprintf("Subtotal: $%.2f", m.invoice.Subtotal.InexactFloat64())) + "\n")
		if m.invoice.DiscountRate.GreaterThan(decimal.Zero) {
			s.WriteString(fmt.Sprintf("%*s", 74, fmt.Sprintf("Discount: -$%.2f", m.invoice.Discount.InexactFloat64())) + "\n")
		}
		if m.invoice.TaxRate.GreaterThan(decimal.Zero) {
			s.WriteString(fmt.Sprintf("%*s", 74, fmt.Sprintf("Tax: $%.2f", m.invoice.Tax.InexactFloat64())) + "\n")
		}
		s.WriteString(fmt.Sprintf("%*s", 74, fmt.Sprintf("Total: $%.2f", m.invoice.Total.InexactFloat64())) + "\n")
	}
	
	s.WriteString("\n" + helpStyle.Render("a add • d delete • s save • ↑/k up • ↓/j down • esc back"))
	
	return appStyle.Render(s.String())
}

func (m InvoiceFormModel) viewAddLineItem() string {
	var s strings.Builder
	
	s.WriteString(titleStyle.Render("Add Line Item") + "\n\n")
	
	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
		m.err = nil
	}
	
	labels := []string{"Description:", "Quantity:", "Unit Price:"}
	for i, label := range labels {
		s.WriteString(formLabelStyle.Render(label))
		s.WriteString(m.lineItemInputs[i].View())
		s.WriteString("\n")
	}
	
	s.WriteString("\n")
	
	button := "[ Add Item ]"
	if m.lineItemFocusIndex == len(m.lineItemInputs) {
		button = selectedStyle.Render(button)
	}
	s.WriteString(button)
	
	s.WriteString("\n\n" + helpStyle.Render("tab navigate • enter add • esc cancel"))
	
	return appStyle.Render(s.String())
}