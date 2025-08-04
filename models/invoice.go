package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InvoiceStatus string

const (
	StatusDraft InvoiceStatus = "draft"
	StatusSent  InvoiceStatus = "sent"
	StatusPaid  InvoiceStatus = "paid"
	StatusOverdue InvoiceStatus = "overdue"
)

type LineItem struct {
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Quantity    decimal.Decimal `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Total       decimal.Decimal `json:"total"`
}

func NewLineItem(description string, quantity, unitPrice decimal.Decimal) *LineItem {
	return &LineItem{
		ID:          uuid.New().String(),
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		Total:       quantity.Mul(unitPrice),
	}
}

func (li *LineItem) UpdateTotal() {
	li.Total = li.Quantity.Mul(li.UnitPrice)
}

type Invoice struct {
	ID           string          `json:"id"`
	Number       string          `json:"number"`
	ClientID     string          `json:"client_id"`
	ClientName   string          `json:"client_name"`
	Date         time.Time       `json:"date"`
	DueDate      time.Time       `json:"due_date"`
	LineItems    []LineItem      `json:"line_items"`
	Subtotal     decimal.Decimal `json:"subtotal"`
	DiscountRate decimal.Decimal `json:"discount_rate"`
	Discount     decimal.Decimal `json:"discount"`
	TaxRate      decimal.Decimal `json:"tax_rate"`
	Tax          decimal.Decimal `json:"tax"`
	Total        decimal.Decimal `json:"total"`
	Status       InvoiceStatus   `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

func NewInvoice(clientID, clientName, number string) *Invoice {
	now := time.Now()
	return &Invoice{
		ID:           uuid.New().String(),
		Number:       number,
		ClientID:     clientID,
		ClientName:   clientName,
		Date:         now,
		DueDate:      now.AddDate(0, 0, 30),
		LineItems:    []LineItem{},
		Subtotal:     decimal.Zero,
		DiscountRate: decimal.Zero,
		Discount:     decimal.Zero,
		TaxRate:      decimal.Zero,
		Tax:          decimal.Zero,
		Total:        decimal.Zero,
		Status:       StatusDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (i *Invoice) AddLineItem(item LineItem) {
	i.LineItems = append(i.LineItems, item)
	i.CalculateTotals()
}

func (i *Invoice) RemoveLineItem(itemID string) {
	newItems := []LineItem{}
	for _, item := range i.LineItems {
		if item.ID != itemID {
			newItems = append(newItems, item)
		}
	}
	i.LineItems = newItems
	i.CalculateTotals()
}

func (i *Invoice) CalculateTotals() {
	subtotal := decimal.Zero
	for _, item := range i.LineItems {
		subtotal = subtotal.Add(item.Total)
	}
	i.Subtotal = subtotal
	
	if i.DiscountRate.GreaterThan(decimal.Zero) {
		i.Discount = i.Subtotal.Mul(i.DiscountRate.Div(decimal.NewFromInt(100)))
	} else {
		i.Discount = decimal.Zero
	}
	
	afterDiscount := i.Subtotal.Sub(i.Discount)
	
	if i.TaxRate.GreaterThan(decimal.Zero) {
		i.Tax = afterDiscount.Mul(i.TaxRate.Div(decimal.NewFromInt(100)))
	} else {
		i.Tax = decimal.Zero
	}
	
	i.Total = afterDiscount.Add(i.Tax)
	i.UpdatedAt = time.Now()
}

var DecimalZero = decimal.Zero

func (i *Invoice) SetDiscountRate(rate decimal.Decimal) {
	i.DiscountRate = rate
	i.CalculateTotals()
}

func (i *Invoice) SetTaxRate(rate decimal.Decimal) {
	i.TaxRate = rate
	i.CalculateTotals()
}

func (i *Invoice) UpdateStatus(newStatus InvoiceStatus, reason string) error {
	if i.Status == newStatus {
		return fmt.Errorf("invoice already has status %s", newStatus)
	}
	
	i.Status = newStatus
	i.UpdatedAt = time.Now()
	return nil
}

func GenerateInvoiceNumber(year int, sequence int) string {
	return fmt.Sprintf("%d-%02d", year, sequence)
}