package models

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewClient(name, address, email string) *Client {
	now := time.Now()
	return &Client{
		ID:        uuid.New().String(),
		Name:      name,
		Address:   address,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Client) Update(name, address, email string) {
	c.Name = name
	c.Address = address
	c.Email = email
	c.UpdatedAt = time.Now()
}