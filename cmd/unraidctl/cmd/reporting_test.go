package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/jwmoss/unraidctl/internal/output"
	"github.com/jwmoss/unraidctl/pkg/client"
)

func captureOutput(t *testing.T, run func() error) (string, error) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout, oldOut := os.Stdout, out
	os.Stdout = writer
	out = output.New(false, false, true)
	defer func() { os.Stdout, out = oldStdout, oldOut }()
	result := make(chan string, 1)
	go func() { data, _ := io.ReadAll(reader); result <- string(data) }()
	runErr := run()
	_ = writer.Close()
	text := <-result
	_ = reader.Close()
	return text, runErr
}

func useResponse(t *testing.T, response string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, response)
	}))
	old := apiClient
	apiClient = client.New(server.URL, "test-key")
	t.Cleanup(func() { apiClient = old; server.Close() })
}

func TestCapacityUnits(t *testing.T) {
	if got := formatSizeKB(97990693745); got != "98.0 TB" {
		t.Errorf("decimal KB: %s", got)
	}
	if got := formatSizeBytes(1024); got != "1.0 KiB" {
		t.Errorf("binary bytes: %s", got)
	}
	var arr api.Array
	if err := json.Unmarshal([]byte(`{"state":"STARTED","capacity":{"kilobytes":{"total":"97990693745","used":"11441878941","free":"86548814804"}}}`), &arr); err != nil {
		t.Fatal(err)
	}
	text, _ := captureOutput(t, func() error { renderArray(arr); return nil })
	if !strings.Contains(text, "97.99 TB") {
		t.Fatalf("wrong capacity: %s", text)
	}
}

func TestArrayIncludesParityCacheAndBoot(t *testing.T) {
	var arr api.Array
	if err := json.Unmarshal([]byte(`{"state":"STARTED","parities":[{"name":"parity","type":"PARITY","size":13672382412}],"caches":[{"name":"cache","type":"CACHE"}],"boot":{"name":"flash","type":"FLASH"}}`), &arr); err != nil {
		t.Fatal(err)
	}
	text, _ := captureOutput(t, func() error { renderArray(arr); return nil })
	for _, want := range []string{"parity", "cache", "flash", "12.73 TiB"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q: %s", want, text)
		}
	}
}

func TestInfoUsesExactVersions(t *testing.T) {
	useResponse(t, `{"data":{"info":{"os":{"distro":"Unraid OS","release":"7.3 x86_64"},"versions":{"core":{"unraid":"7.3.2","api":"4.37.4","kernel":"6.18.38-Unraid"}}}}}`)
	text, err := captureOutput(t, func() error { return infoCmd.RunE(infoCmd, nil) })
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"7.3.2", "4.37.4", "6.18.38-Unraid"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q: %s", want, text)
		}
	}
}

func TestRetiredDiskRemovalMakesNoRequest(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, _ = io.WriteString(w, `{"data":{"array":{"removeDiskFromArray":{}}}}`)
	}))
	defer server.Close()
	old := apiClient
	apiClient = client.New(server.URL, "test-key")
	defer func() { apiClient = old }()
	_, err := captureOutput(t, func() error { return arrayRemoveDiskCmd.RunE(arrayRemoveDiskCmd, []string{"disk-id"}) })
	if called {
		t.Error("retired command sent a mutation")
	}
	if err == nil || !strings.Contains(err.Error(), "Unraid WebGUI") {
		t.Fatalf("expected supported workflow guidance, got %v", err)
	}
}

func TestVMUnavailableIsExplicit(t *testing.T) {
	useResponse(t, `{"errors":[{"message":"Failed to retrieve VM domains: VMs are not available"}]}`)
	_, err := captureOutput(t, func() error { return vmListCmd.RunE(vmListCmd, nil) })
	if err == nil || !strings.Contains(err.Error(), "VM Manager is disabled or unavailable") {
		t.Fatalf("unhelpful VM error: %v", err)
	}
}
