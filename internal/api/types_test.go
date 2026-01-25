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

func TestArrayResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"array": {
			"state": "STARTED",
			"capacity": {
				"kilobytes": {
					"free": "29592689497",
					"used": "68398004248",
					"total": "97990693745"
				}
			},
			"disks": [
				{
					"id": "disk1-id",
					"name": "disk1",
					"device": "sdd",
					"size": 13672382412,
					"status": "DISK_OK",
					"temp": 34,
					"type": "DATA"
				}
			]
		}
	}`

	var resp ArrayResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Array.State != "STARTED" {
		t.Errorf("expected state STARTED, got %s", resp.Array.State)
	}
	if resp.Array.Capacity.Kilobytes.Total != "97990693745" {
		t.Errorf("expected total 97990693745, got %s", resp.Array.Capacity.Kilobytes.Total)
	}
	if len(resp.Array.Disks) != 1 {
		t.Fatalf("expected 1 disk, got %d", len(resp.Array.Disks))
	}
	if resp.Array.Disks[0].Name != "disk1" {
		t.Errorf("expected disk name disk1, got %s", resp.Array.Disks[0].Name)
	}
	if resp.Array.Disks[0].Temp != 34 {
		t.Errorf("expected temp 34, got %d", resp.Array.Disks[0].Temp)
	}
}

func TestDockerResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"docker": {
			"containers": [
				{
					"id": "abc123",
					"names": ["/plex"],
					"state": "RUNNING",
					"status": "Up 16 hours",
					"image": "linuxserver/plex",
					"autoStart": true
				},
				{
					"id": "def456",
					"names": ["/sonarr"],
					"state": "EXITED",
					"status": "Exited (0) 2 hours ago",
					"image": "linuxserver/sonarr",
					"autoStart": false
				}
			]
		}
	}`

	var resp DockerResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.Docker.Containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(resp.Docker.Containers))
	}

	plex := resp.Docker.Containers[0]
	if plex.Names[0] != "/plex" {
		t.Errorf("expected name /plex, got %s", plex.Names[0])
	}
	if plex.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", plex.State)
	}
	if !plex.AutoStart {
		t.Error("expected autoStart true")
	}

	sonarr := resp.Docker.Containers[1]
	if sonarr.AutoStart {
		t.Error("expected autoStart false")
	}
}

func TestSharesResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"shares": [
			{
				"name": "appdata",
				"comment": "application data",
				"free": 938371584,
				"used": 83130216
			},
			{
				"name": "media",
				"comment": "",
				"free": 29592689496,
				"used": 68398004249
			}
		]
	}`

	var resp SharesResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.Shares) != 2 {
		t.Fatalf("expected 2 shares, got %d", len(resp.Shares))
	}

	appdata := resp.Shares[0]
	if appdata.Name != "appdata" {
		t.Errorf("expected name appdata, got %s", appdata.Name)
	}
	if appdata.Comment != "application data" {
		t.Errorf("expected comment 'application data', got %s", appdata.Comment)
	}
	if appdata.Free != 938371584 {
		t.Errorf("expected free 938371584, got %d", appdata.Free)
	}
}

func TestNotificationsResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"notifications": {
			"overview": {
				"unread": {
					"total": 5
				}
			},
			"list": [
				{
					"id": "notif-1",
					"subject": "Parity check complete",
					"importance": "WARNING",
					"timestamp": "2026-01-16T15:15:01.000Z"
				},
				{
					"id": "notif-2",
					"subject": "Docker update available",
					"importance": "INFO",
					"timestamp": "2026-01-17T10:00:00.000Z"
				}
			]
		}
	}`

	var resp NotificationsResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Notifications.Overview.Unread.Total != 5 {
		t.Errorf("expected 5 unread, got %d", resp.Notifications.Overview.Unread.Total)
	}

	if len(resp.Notifications.List) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(resp.Notifications.List))
	}

	notif := resp.Notifications.List[0]
	if notif.Subject != "Parity check complete" {
		t.Errorf("expected subject 'Parity check complete', got %s", notif.Subject)
	}
	if notif.Importance != "WARNING" {
		t.Errorf("expected importance WARNING, got %s", notif.Importance)
	}
}

func TestVMsResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"vms": {
			"id": "vms-id",
			"domain": [
				{
					"name": "Windows 11",
					"state": "running"
				},
				{
					"name": "Ubuntu Server",
					"state": "shutoff"
				}
			]
		}
	}`

	var resp VMsResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.VMs.Domain) != 2 {
		t.Fatalf("expected 2 VMs, got %d", len(resp.VMs.Domain))
	}

	win := resp.VMs.Domain[0]
	if win.Name != "Windows 11" {
		t.Errorf("expected name 'Windows 11', got %s", win.Name)
	}
	if win.State != "running" {
		t.Errorf("expected state running, got %s", win.State)
	}
}
