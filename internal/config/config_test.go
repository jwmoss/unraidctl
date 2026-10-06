package config

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
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

	// Set environment variables (t.Setenv auto-cleans up after test)
	t.Setenv("UNRAID_SERVER", "http://env-server.local")
	t.Setenv("UNRAID_API_KEY", "env-api-key")

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
	// Set environment variables (t.Setenv auto-cleans up after test)
	t.Setenv("UNRAID_SERVER", "http://env-only-server.local")
	t.Setenv("UNRAID_API_KEY", "env-only-api-key")

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

func TestLoad_ReadErrors(t *testing.T) {
	t.Setenv("UNRAID_SERVER", "http://env-server.local")
	t.Setenv("UNRAID_API_KEY", "synthetic-env-key")
	for _, name := range []string{"directory", "symlink loop", "permission denied"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			switch name {
			case "directory":
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink loop":
				if err := os.Symlink(path, path); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			case "permission denied":
				if runtime.GOOS == "windows" || os.Geteuid() == 0 {
					t.Skip("file permissions require an unprivileged Unix user")
				}
				if err := os.WriteFile(path, []byte("server: http://file.invalid\n"), 0000); err != nil {
					t.Fatal(err)
				}
			}
			_, readErr := os.ReadFile(path)
			var pathErr *os.PathError
			if !errors.As(readErr, &pathErr) || os.IsNotExist(readErr) {
				t.Fatalf("fixture must cause a read error: %v", readErr)
			}
			cfg, err := Load(path)
			if cfg != nil || !errors.Is(err, pathErr.Err) || !errors.As(err, &pathErr) {
				t.Fatalf("Load() = %v, %v; want wrapped read error %v", cfg, err, readErr)
			}
		})
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
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("config permissions = %o; want 600", info.Mode().Perm())
	}
}

func TestSave_PrivateOverwrite(t *testing.T) {
	t.Setenv("UNRAID_SERVER", "")
	t.Setenv("UNRAID_API_KEY", "")
	path := filepath.Join(t.TempDir(), "config.yaml")
	original := []byte("server: http://old.invalid\napi_key: synthetic-old-key\n")
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(filepath.Dir(path), "old.yaml")
	if err := os.Link(path, alias); err != nil {
		t.Skipf("hard links unavailable: %v", err)
	}
	want := &Config{Server: "http://new.invalid", APIKey: "synthetic-new-key"}
	if err := Save(want, path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Errorf("config permissions = %o; want 600", info.Mode().Perm())
	}
	got, err := Load(path)
	if err != nil || got == nil || *got != *want {
		t.Errorf("Load() after overwrite = %v, %v; want %v", got, err, want)
	}
	data, err := os.ReadFile(alias)
	if err != nil || !bytes.Equal(data, original) {
		t.Errorf("Save() changed the old file through a hard link: %q, %v", data, err)
	}
}

func TestSave_RejectsNonregularDestination(t *testing.T) {
	for _, name := range []string{"symlink", "dangling symlink", "directory"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path, target := filepath.Join(dir, "config.yaml"), filepath.Join(dir, "target.yaml")
			original := []byte("original config bytes\n")
			if name == "directory" {
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			} else {
				if name == "symlink" {
					if err := os.WriteFile(target, original, 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(target, path); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			}
			if err := Save(&Config{Server: "http://new.invalid", APIKey: "synthetic-new-key"}, path); err == nil {
				t.Error("Save() accepts a nonregular destination")
			}
			info, err := os.Lstat(path)
			if err != nil || (name == "directory" && !info.IsDir()) || (name != "directory" && info.Mode()&os.ModeSymlink == 0) {
				t.Fatalf("Save() changes the destination: %v, %v", info, err)
			}
			data, err := os.ReadFile(target)
			if name == "symlink" && (err != nil || !bytes.Equal(data, original)) {
				t.Errorf("Save() changes the symlink target: %q, %v", data, err)
			}
			if name == "dangling symlink" && !os.IsNotExist(err) {
				t.Errorf("Save() creates the symlink target: %v", err)
			}
		})
	}
}

func TestSave_UnwritableDirectoryPreservesFile(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("directory permissions require an unprivileged Unix user")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	original := []byte("original config bytes\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(dir, 0700); err != nil {
			t.Error(err)
		}
	})
	err := Save(&Config{Server: "http://new.invalid", APIKey: "synthetic-new-key"}, path)
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("Save() = %v; want permission error", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, original) {
		t.Errorf("Save() changes old contents on failure: %q, %v", data, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "config.yaml" {
		t.Errorf("Save() leaves temporary files: %v, %v", entries, err)
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
