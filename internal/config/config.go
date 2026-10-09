// Package config persists user settings in ~/.stremcli.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	dirName  = ".stremcli"
	fileName = "config.json"
)

// Config holds persisted user settings.
type Config struct {
	RealDebridToken string `json:"real_debrid_token,omitempty"`
	// Player is used when no -player flag is given.
	Player string `json:"player,omitempty"`
}

// Dir returns the stremcli data directory (~/.stremcli).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(home, dirName), nil
}

// EnsureDir returns the data directory, creating it if needed.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}

// Path returns the location of the config file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Load reads the config file. A missing file yields an empty Config.
func Load() (Config, error) {
	var cfg Config
	p, err := Path()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", p, err)
	}
	return cfg, nil
}

// Save writes the config file with owner-only permissions, since it holds
// an API token.
func Save(cfg Config) (string, error) {
	dir, err := EnsureDir()
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}

	p := filepath.Join(dir, fileName)
	if err := os.WriteFile(p, append(data, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}
	return p, nil
}
