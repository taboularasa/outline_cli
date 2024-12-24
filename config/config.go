package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey     string `json:"api_key"`
	OutlineURL string `json:"outline_url"`
}

var LoadConfig = loadConfig

func loadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".outline-cli", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
