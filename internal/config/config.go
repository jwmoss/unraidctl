package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server string `yaml:"server"`
	APIKey string `yaml:"api_key"`
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "unraidctl", "config.yaml")
}

func Load(path string) (*Config, error) {
	cfg := &Config{}

	// Try to load from file
	if path == "" {
		path = DefaultConfigPath()
	}

	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Override with environment variables
	if server := os.Getenv("UNRAID_SERVER"); server != "" {
		cfg.Server = server
	}
	if apiKey := os.Getenv("UNRAID_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server == "" {
		return fmt.Errorf("server URL is required (set via --server, UNRAID_SERVER, or config file)")
	}
	if c.APIKey == "" {
		return fmt.Errorf("API key is required (set via --api-key, UNRAID_API_KEY, or config file)")
	}
	return nil
}

func EnsureConfigDir() error {
	path := DefaultConfigPath()
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0700)
}

func Save(cfg *Config, path string) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	// Ensure the directory exists for the given path
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("refusing non-regular config path %s", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect config file: %w", err)
	}

	file, err := os.CreateTemp(dir, ".config-*")
	if err != nil {
		return fmt.Errorf("failed to create private config file: %w", err)
	}
	defer func() { _ = os.Remove(file.Name()) }()
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to write config file: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to sync config file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close config file: %w", err)
	}
	// Replace the directory entry. Do not follow a symlink created after Lstat.
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("failed to publish config file: %w", err)
	}
	return nil
}
