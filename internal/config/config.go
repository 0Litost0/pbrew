package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/pbrew/internal/model"
)

const (
	configDir  = ".pbrew"
	configFile = "config.json"
)

// configPath returns the absolute path to the config file (~/.pbrew/config.json).
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load reads the configuration from disk.
// Returns a zero-value Config if the file does not exist.
func Load() (*model.Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.Config{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg model.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Save writes the configuration to disk using atomic write-temp-then-rename.
func (c *modelConfigWrapper) Save() error {
	return save(c.cfg)
}

type modelConfigWrapper struct {
	cfg *model.Config
}

// Wrap creates a saveable wrapper around a Config.
func Wrap(cfg *model.Config) *modelConfigWrapper {
	return &modelConfigWrapper{cfg: cfg}
}

func save(cfg *model.Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return atomicWriteFile(path, data, 0644)
}

// Get returns the value for the given top-level key.
func Get(cfg *model.Config, key string) (string, bool) {
	switch key {
	case "source":
		return cfg.Source, cfg.Source != ""
	default:
		return "", false
	}
}

// Set sets the value for the given top-level key and saves to disk.
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "source":
		cfg.Source = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return save(cfg)
}

// atomicWriteFile writes data to path atomically using write-temp-then-rename.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".pbrew-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}
