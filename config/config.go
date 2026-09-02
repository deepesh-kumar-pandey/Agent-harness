package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Provider Provider `json:"provider"`
}

type Provider struct {
	Name     string `json:"name"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
	Endpoint string `json:"endpoint"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Provider.Name == "" {
		return fmt.Errorf("provider name is required")
	}

	if c.Provider.Model == "" {
		return fmt.Errorf("provider model is required")
	}

	if c.Provider.BaseURL == "" {
		return fmt.Errorf("provider base URL is required")
	}

	if c.Provider.Endpoint == "" {
		return fmt.Errorf("provider endpoint is required")
	}

	return nil
}
