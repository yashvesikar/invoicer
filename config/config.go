package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DataPath       string `json:"data_path"`
	CompanyName    string `json:"company_name"`
	CompanyAddress string `json:"company_address"`
	CompanyEmail   string `json:"company_email"`
	// Payment information
	ZelleAccount   string `json:"zelle_account,omitempty"`
	VenmoAccount   string `json:"venmo_account,omitempty"`
	BankName       string `json:"bank_name,omitempty"`
	BankRouting    string `json:"bank_routing,omitempty"`
	BankAccount    string `json:"bank_account,omitempty"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		DataPath:       filepath.Join(homeDir, ".config", "invoicer"),
		CompanyName:    "Your Company Name",
		CompanyAddress: "123 Main St, City, Country",
		CompanyEmail:   "billing@yourcompany.com",
	}
}

func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "invoicer", "config.json"), nil
}

func Load() (*Config, error) {
	// First check environment variable
	if dataPath := os.Getenv("INVOICER_DATA_PATH"); dataPath != "" {
		config := DefaultConfig()
		config.DataPath = dataPath
		return config, nil
	}

	// Then check config file
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist, will need setup
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) DataDir() string {
	return filepath.Join(c.DataPath, "data")
}

func (c *Config) TemplatesDir() string {
	return filepath.Join(c.DataPath, "templates")
}

func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.DataPath,
		c.DataDir(),
		c.TemplatesDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Check if template exists, if not copy from the working directory
	templatePath := filepath.Join(c.TemplatesDir(), "invoice.tex")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Try to copy from local templates directory
		localTemplate := "./templates/invoice.tex"
		if _, err := os.Stat(localTemplate); err == nil {
			// Copy the template
			if err := copyTemplateFile(localTemplate, templatePath); err != nil {
				return fmt.Errorf("failed to copy template: %w", err)
			}
		}
	}

	return nil
}

func copyTemplateFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}