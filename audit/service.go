package audit

import (
	"github.com/user/invoicer/models"
)

type Service struct {
	storage models.Storage
}

func NewService(storage models.Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) LogStatusChange(invoice *models.Invoice, oldStatus models.InvoiceStatus, reason string) error {
	entry := models.NewAuditEntry(
		invoice.ID,
		invoice.Number,
		oldStatus,
		invoice.Status,
		reason,
	)
	
	return s.storage.SaveAuditEntry(entry)
}