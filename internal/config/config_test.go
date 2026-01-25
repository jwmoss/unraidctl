package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: "http://192.168.1.100",
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "missing server",
			config: &Config{
				Server: "",
				APIKey: "test-api-key",
			},
			wantErr: true,
		},
		{
			name: "missing api key",
			config: &Config{
				Server: "http://192.168.1.100",
				APIKey: "",
			},
			wantErr: true,
		},
		{
			name: "both missing",
			config: &Config{
				Server: "",
				APIKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad_FromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `server: http://test-server.local
api_key: file-api-key
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server != "http://test-server.local" {
		t.Errorf("expected server http://test-server.local, got %s", cfg.Server)
	}
	if cfg.APIKey != "file-api-key" {
		t.Errorf("expected api_key file-api-key, got %s", cfg.APIKey)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `server: http://file-server.local
api_key: file-api-key
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set environment variables
	os.Setenv("UNRAID_SERVER", "http://env-server.local")
	os.Setenv("UNRAID_API_KEY", "env-api-key")
	defer func() {
		os.Unsetenv("UNRAID_SERVER")
		os.Unsetenv("UNRAID_API_KEY")
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Environment should override file
	if cfg.Server != "http://env-server.local" {
		t.Errorf("expected server from env, got %s", cfg.Server)
	}
	if cfg.APIKey != "env-api-key" {
		t.Errorf("expected api_key from env, got %s", cfg.APIKey)
	}
}

func TestLoad_EnvOnly(t *testing.T) {
	// Set environment variables
	os.Setenv("UNRAID_SERVER", "http://env-only-server.local")
	os.Setenv("UNRAID_API_KEY", "env-only-api-key")
	defer func() {
		os.Unsetenv("UNRAID_SERVER")
		os.Unsetenv("UNRAID_API_KEY")
	}()

	// Load with non-existent config file
	cfg, err := Load("/non/existent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server != "http://env-only-server.local" {
		t.Errorf("expected server from env, got %s", cfg.Server)
	}
	if cfg.APIKey != "env-only-api-key" {
		t.Errorf("expected api_key from env, got %s", cfg.APIKey)
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := &Config{
		Server: "http://save-test.local",
		APIKey: "save-test-key",
	}

	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created and is readable
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.Server != cfg.Server {
		t.Errorf("expected server %s, got %s", cfg.Server, loaded.Server)
	}
	if loaded.APIKey != cfg.APIKey {
		t.Errorf("expected api_key %s, got %s", cfg.APIKey, loaded.APIKey)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Error("DefaultConfigPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultConfigPath() should return absolute path, got %s", path)
	}
	if !contains(path, "unraidctl") {
		t.Errorf("DefaultConfigPath() should contain 'unraidctl', got %s", path)
	}
}

func contains(s, substr string) bool {
	return filepath.Base(filepath.Dir(s)) == "unraidctl" || filepath.Base(s) == "config.yaml"
}
