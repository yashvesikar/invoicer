package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditEntry struct {
	ID            string        `json:"id"`
	InvoiceID     string        `json:"invoice_id"`
	InvoiceNumber string        `json:"invoice_number"`
	OldStatus     InvoiceStatus `json:"old_status"`
	NewStatus     InvoiceStatus `json:"new_status"`
	ChangedBy     string        `json:"changed_by"`
	ChangedAt     time.Time     `json:"changed_at"`
	Reason        string        `json:"reason,omitempty"`
}

func NewAuditEntry(invoiceID, invoiceNumber string, oldStatus, newStatus InvoiceStatus, reason string) *AuditEntry {
	return &AuditEntry{
		ID:            uuid.New().String(),
		InvoiceID:     invoiceID,
		InvoiceNumber: invoiceNumber,
		OldStatus:     oldStatus,
		NewStatus:     newStatus,
		ChangedBy:     "system",
		ChangedAt:     time.Now(),
		Reason:        reason,
	}
}