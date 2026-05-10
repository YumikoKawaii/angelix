package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ServerURL   string `json:"server_url"`
	MemberToken string `json:"member_token"`
}

// IsNotConfigured reports whether err came from a missing config file.
func IsNotConfigured(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

func LoadConfig() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config not found at %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("server_url is required in config")
	}
	if cfg.MemberToken == "" {
		return nil, fmt.Errorf("member_token is required in config")
	}
	return &cfg, nil
}

func configPath() string {
	if v := os.Getenv("ANGELIX_CONFIG"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".angelix", "config.json")
}
