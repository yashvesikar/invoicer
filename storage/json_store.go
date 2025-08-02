package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/user/invoicer/models"
)

type JSONStorage struct {
	dataDir      string
	clientsFile  string
	invoicesFile string
	mu           sync.RWMutex
}

func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	s := &JSONStorage{
		dataDir:      dataDir,
		clientsFile:  filepath.Join(dataDir, "clients.json"),
		invoicesFile: filepath.Join(dataDir, "invoices.json"),
	}
	
	if err := s.initFiles(); err != nil {
		return nil, err
	}
	
	return s, nil
}

func (s *JSONStorage) initFiles() error {
	files := []string{s.clientsFile, s.invoicesFile}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if err := os.WriteFile(file, []byte("[]"), 0644); err != nil {
				return fmt.Errorf("failed to create %s: %w", file, err)
			}
		}
	}
	return nil
}

func (s *JSONStorage) readClients() ([]models.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	data, err := os.ReadFile(s.clientsFile)
	if err != nil {
		return nil, err
	}
	
	var clients []models.Client
	if err := json.Unmarshal(data, &clients); err != nil {
		return nil, err
	}
	
	return clients, nil
}

func (s *JSONStorage) writeClients(clients []models.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	data, err := json.MarshalIndent(clients, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(s.clientsFile, data, 0644)
}

func (s *JSONStorage) readInvoices() ([]models.Invoice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	data, err := os.ReadFile(s.invoicesFile)
	if err != nil {
		return nil, err
	}
	
	var invoices []models.Invoice
	if err := json.Unmarshal(data, &invoices); err != nil {
		return nil, err
	}
	
	return invoices, nil
}

func (s *JSONStorage) writeInvoices(invoices []models.Invoice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	data, err := json.MarshalIndent(invoices, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(s.invoicesFile, data, 0644)
}

func (s *JSONStorage) GetAllClients() ([]models.Client, error) {
	return s.readClients()
}

func (s *JSONStorage) GetClient(id string) (*models.Client, error) {
	clients, err := s.readClients()
	if err != nil {
		return nil, err
	}
	
	for _, client := range clients {
		if client.ID == id {
			return &client, nil
		}
	}
	
	return nil, errors.New("client not found")
}

func (s *JSONStorage) SaveClient(client *models.Client) error {
	clients, err := s.readClients()
	if err != nil {
		return err
	}
	
	clients = append(clients, *client)
	return s.writeClients(clients)
}

func (s *JSONStorage) UpdateClient(client *models.Client) error {
	clients, err := s.readClients()
	if err != nil {
		return err
	}
	
	for i, c := range clients {
		if c.ID == client.ID {
			clients[i] = *client
			return s.writeClients(clients)
		}
	}
	
	return errors.New("client not found")
}

func (s *JSONStorage) DeleteClient(id string) error {
	clients, err := s.readClients()
	if err != nil {
		return err
	}
	
	newClients := []models.Client{}
	found := false
	for _, c := range clients {
		if c.ID != id {
			newClients = append(newClients, c)
		} else {
			found = true
		}
	}
	
	if !found {
		return errors.New("client not found")
	}
	
	return s.writeClients(newClients)
}

func (s *JSONStorage) GetAllInvoices() ([]models.Invoice, error) {
	return s.readInvoices()
}

func (s *JSONStorage) GetInvoice(id string) (*models.Invoice, error) {
	invoices, err := s.readInvoices()
	if err != nil {
		return nil, err
	}
	
	for _, invoice := range invoices {
		if invoice.ID == id {
			return &invoice, nil
		}
	}
	
	return nil, errors.New("invoice not found")
}

func (s *JSONStorage) GetInvoiceByNumber(number string) (*models.Invoice, error) {
	invoices, err := s.readInvoices()
	if err != nil {
		return nil, err
	}
	
	for _, invoice := range invoices {
		if invoice.Number == number {
			return &invoice, nil
		}
	}
	
	return nil, errors.New("invoice not found")
}

func (s *JSONStorage) SaveInvoice(invoice *models.Invoice) error {
	invoices, err := s.readInvoices()
	if err != nil {
		return err
	}
	
	invoices = append(invoices, *invoice)
	return s.writeInvoices(invoices)
}

func (s *JSONStorage) UpdateInvoice(invoice *models.Invoice) error {
	invoices, err := s.readInvoices()
	if err != nil {
		return err
	}
	
	for i, inv := range invoices {
		if inv.ID == invoice.ID {
			invoices[i] = *invoice
			return s.writeInvoices(invoices)
		}
	}
	
	return errors.New("invoice not found")
}

func (s *JSONStorage) DeleteInvoice(id string) error {
	invoices, err := s.readInvoices()
	if err != nil {
		return err
	}
	
	newInvoices := []models.Invoice{}
	found := false
	for _, inv := range invoices {
		if inv.ID != id {
			newInvoices = append(newInvoices, inv)
		} else {
			found = true
		}
	}
	
	if !found {
		return errors.New("invoice not found")
	}
	
	return s.writeInvoices(newInvoices)
}

func (s *JSONStorage) GetNextInvoiceNumber(year int) (int, error) {
	invoices, err := s.readInvoices()
	if err != nil {
		return 0, err
	}
	
	maxSequence := 0
	yearPrefix := fmt.Sprintf("%d-", year)
	
	for _, invoice := range invoices {
		if len(invoice.Number) >= 7 && invoice.Number[:5] == yearPrefix {
			var seq int
			if _, err := fmt.Sscanf(invoice.Number[5:], "%d", &seq); err == nil {
				if seq > maxSequence {
					maxSequence = seq
				}
			}
		}
	}
	
	return maxSequence + 1, nil
}

func (s *JSONStorage) GetInvoicesByClient(clientID string) ([]models.Invoice, error) {
	invoices, err := s.readInvoices()
	if err != nil {
		return nil, err
	}
	
	clientInvoices := []models.Invoice{}
	for _, invoice := range invoices {
		if invoice.ClientID == clientID {
			clientInvoices = append(clientInvoices, invoice)
		}
	}
	
	return clientInvoices, nil
}