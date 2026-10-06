package tests

import (
	"archive/zip"
	"context"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVersionBuildMetadata(t *testing.T) {
	// Use a file proxy to test go install without network access or live credentials.
	const module = "github.com/jwmoss/unraidctl"
	const moduleVersion = "v1.2.3"
	root := filepath.Clean("..")
	dir := t.TempDir()
	cache, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	proxy := filepath.Join(dir, "proxy")
	versions := filepath.Join(proxy, filepath.FromSlash(module), "@v")
	if err := os.MkdirAll(versions, 0700); err != nil {
		t.Fatal(err)
	}
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		moduleVersion + ".mod":  mod,
		moduleVersion + ".info": []byte(`{"Version":"v1.2.3","Time":"2026-01-01T00:00:00Z"}`),
		"list":                  []byte(moduleVersion + "\n"),
	} {
		if err := os.WriteFile(filepath.Join(versions, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	archive, err := os.Create(filepath.Join(versions, moduleVersion+".zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = archive.Close() }()
	writer := zip.NewWriter(archive)
	for _, source := range []string{"go.mod", "go.sum", "cmd", "internal", "pkg"} {
		err := filepath.WalkDir(filepath.Join(root, source), func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			if entry.Name() != "go.mod" && entry.Name() != "go.sum" && filepath.Ext(path) != ".go" {
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			file, err := writer.Create(module + "@" + moduleVersion + "/" + filepath.ToSlash(relative))
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			_, err = file.Write(data)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	fileURL := func(path string) string {
		return (&url.URL{Scheme: "file", Path: "/" + strings.TrimPrefix(filepath.ToSlash(path), "/")}).String()
	}
	t.Setenv("GOPROXY", fileURL(proxy)+","+fileURL(filepath.Join(strings.TrimSpace(string(cache)), "cache", "download")))
	t.Setenv("GOMODCACHE", filepath.Join(dir, "modules"))
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GOPRIVATE", "")
	t.Setenv("GONOPROXY", "")
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "-modcacherw")
	t.Setenv("GOBIN", filepath.Join(dir, "bin"))
	for _, tt := range []struct {
		name, flags, want string
		source            bool
	}{
		{"module version", "", "unraidctl version v1.2.3\ncommit: unknown\nbuilt:  unknown\n", false},
		{"explicit build flags", "-X " + module + "/cmd/unraidctl/cmd.version=9.8.7 -X " + module + "/cmd/unraidctl/cmd.commit=fixture-commit -X " + module + "/cmd/unraidctl/cmd.date=fixture-date", "unraidctl version 9.8.7\ncommit: fixture-commit\nbuilt:  fixture-date\n", false},
		{"source development version", "", "unraidctl version dev\ncommit: unknown\nbuilt:  unknown\n", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			binary := filepath.Join(os.Getenv("GOBIN"), "unraidctl")
			if os.PathSeparator == '\\' {
				binary += ".exe"
			}
			build := exec.CommandContext(ctx, "go", "install", "-ldflags", tt.flags, module+"/cmd/unraidctl@"+moduleVersion)
			if tt.source {
				build = exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-o", binary, "./cmd/unraidctl")
				build.Dir = root
			}
			if output, err := build.CombinedOutput(); err != nil {
				t.Fatalf("offline build: %v\n%s", err, output)
			}
			command := exec.CommandContext(ctx, binary, "--config", dir, "version")
			output, err := command.CombinedOutput()
			if err != nil || string(output) != tt.want {
				t.Fatalf("version output = %q, %v; want %q", output, err, tt.want)
			}
		})
	}
}
