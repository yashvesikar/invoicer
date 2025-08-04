package models

type Storage interface {
	GetAllClients() ([]Client, error)
	GetClient(id string) (*Client, error)
	SaveClient(client *Client) error
	UpdateClient(client *Client) error
	DeleteClient(id string) error
	
	GetAllInvoices() ([]Invoice, error)
	GetInvoice(id string) (*Invoice, error)
	GetInvoiceByNumber(number string) (*Invoice, error)
	SaveInvoice(invoice *Invoice) error
	UpdateInvoice(invoice *Invoice) error
	DeleteInvoice(id string) error
	GetNextInvoiceNumber(year int) (int, error)
	GetInvoicesByClient(clientID string) ([]Invoice, error)
	
	SaveAuditEntry(entry *AuditEntry) error
	GetAuditEntries(invoiceID string) ([]AuditEntry, error)
}