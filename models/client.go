package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Client struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Address          string          `json:"address"`
	Emails           []string        `json:"emails"`
	DefaultHourlyRate decimal.Decimal `json:"default_hourly_rate"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func NewClient(name, address string, emails []string, hourlyRate decimal.Decimal) *Client {
	now := time.Now()
	return &Client{
		ID:               uuid.New().String(),
		Name:             name,
		Address:          address,
		Emails:           emails,
		DefaultHourlyRate: hourlyRate,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (c *Client) Update(name, address string, emails []string, hourlyRate decimal.Decimal) {
	c.Name = name
	c.Address = address
	c.Emails = emails
	c.DefaultHourlyRate = hourlyRate
	c.UpdatedAt = time.Now()
}