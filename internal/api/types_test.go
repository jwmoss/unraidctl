package api

import (
	"encoding/json"
	"testing"
)

func TestInfoResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"info": {
			"os": {
				"platform": "linux",
				"distro": "Unraid OS",
				"release": "7.2 x86_64",
				"uptime": "2026-01-16T15:07:42.840Z",
				"hostname": "Tower"
			},
			"cpu": {
				"manufacturer": "Intel",
				"brand": "Xeon E3-1246 v3",
				"cores": 4,
				"speed": 3.5
			}
		}
	}`

	var resp InfoResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Info.OS.Hostname != "Tower" {
		t.Errorf("expected hostname Tower, got %s", resp.Info.OS.Hostname)
	}
	if resp.Info.OS.Platform != "linux" {
		t.Errorf("expected platform linux, got %s", resp.Info.OS.Platform)
	}
	if resp.Info.OS.Distro != "Unraid OS" {
		t.Errorf("expected distro Unraid OS, got %s", resp.Info.OS.Distro)
	}
	if resp.Info.CPU.Cores != 4 {
		t.Errorf("expected 4 cores, got %d", resp.Info.CPU.Cores)
	}
	if resp.Info.CPU.Speed != 3.5 {
		t.Errorf("expected speed 3.5, got %f", resp.Info.CPU.Speed)
	}
}
