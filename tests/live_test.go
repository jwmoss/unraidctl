//go:build live

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"testing"
	"time"
)

// TestLiveReadOnly requires the real provider and the user's existing configuration.
// It never changes provider state and never logs response bodies or credentials.
func TestLiveReadOnly(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		valid func(any) bool
	}{
		{"server information", []string{"info"}, func(v any) bool {
			data, ok := v.(map[string]any)
			if !ok {
				return false
			}
			system, ok := data["os"].(map[string]any)
			if !ok {
				return false
			}
			hostname, ok := system["hostname"].(string)
			return ok && hostname != ""
		}},
		{"array status", []string{"array", "status"}, func(v any) bool {
			data, ok := v.(map[string]any)
			if !ok {
				return false
			}
			state, ok := data["state"].(string)
			return ok && state != ""
		}},
		{"system services", []string{"system", "services"}, nonEmptyList},
		{"system time", []string{"system", "time"}, func(v any) bool {
			data, ok := v.(map[string]any)
			zone, _ := data["timeZone"].(string)
			return ok && zone != ""
		}},
		{"network interfaces", []string{"system", "interfaces"}, nonEmptyList},
		{"CPU subscription", []string{"metrics", "cpu", "--watch", "--count", "1"}, func(v any) bool {
			data, ok := v.(map[string]any)
			_, hasTotal := data["percentTotal"].(float64)
			return ok && hasTotal
		}},
	}
	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			args := append([]string{"--json"}, check.args...)
			data, err := exec.CommandContext(ctx, "../unraidctl", args...).Output()
			if err != nil {
				t.Fatalf("live CLI failed: %v; check provider access and configuration", err)
			}
			var result any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatal("live CLI returned invalid JSON")
			}
			if !check.valid(result) {
				t.Fatal("live response does not satisfy the command contract")
			}
		})
	}
}

func nonEmptyList(v any) bool {
	list, ok := v.([]any)
	return ok && len(list) > 0
}

func TestLiveServiceInventory(t *testing.T) {
	for _, service := range []struct{ command, unavailable string }{{"docker", "Docker service is unavailable"}, {"vm", "VM Manager is disabled or unavailable"}} {
		t.Run(service.command, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "../unraidctl", "--dry-run", "--json", service.command, "list")
			var diagnostic bytes.Buffer
			cmd.Stderr = &diagnostic
			data, err := cmd.Output()
			if err != nil {
				if len(data) == 0 && bytes.Contains(diagnostic.Bytes(), []byte(service.unavailable)) {
					return
				}
				t.Fatalf("inventory did not return a list or a service-unavailable error: %v", err)
			}
			var items []json.RawMessage
			if json.Unmarshal(data, &items) != nil || items == nil {
				t.Fatal("expected an inventory array")
			}
		})
	}
}
