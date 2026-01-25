package api

import (
	"context"
	"testing"
	"time"

	"github.com/jwmoss/unraidctl/pkg/client"
)

// TestIntegration_InfoQuery tests the info query against the mock Unraid API
func TestIntegration_InfoQuery(t *testing.T) {
	server := MockUnraidAPI()
	defer server.Close()

	c := client.New(server.URL, "test-api-key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp InfoResponse
	err := c.Query(ctx, InfoQuery, nil, &resp)
	if err != nil {
		t.Fatalf("InfoQuery failed: %v", err)
	}

	// Verify response
	if resp.Info.OS.Hostname != "Tower" {
		t.Errorf("expected hostname Tower, got %s", resp.Info.OS.Hostname)
	}
	if resp.Info.OS.Platform != "linux" {
		t.Errorf("expected platform linux, got %s", resp.Info.OS.Platform)
	}
	if resp.Info.OS.Distro != "Unraid OS" {
		t.Errorf("expected distro 'Unraid OS', got %s", resp.Info.OS.Distro)
	}
	if resp.Info.CPU.Manufacturer != "Intel" {
		t.Errorf("expected CPU manufacturer Intel, got %s", resp.Info.CPU.Manufacturer)
	}
	if resp.Info.CPU.Cores != 4 {
		t.Errorf("expected 4 cores, got %d", resp.Info.CPU.Cores)
	}
}

// TestIntegration_ArrayStatusQuery tests the array status query against the mock Unraid API
func TestIntegration_ArrayStatusQuery(t *testing.T) {
	server := MockUnraidAPI()
	defer server.Close()

	c := client.New(server.URL, "test-api-key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp ArrayResponse
	err := c.Query(ctx, ArrayStatusQuery, nil, &resp)
	if err != nil {
		t.Fatalf("ArrayStatusQuery failed: %v", err)
	}

	// Verify response
	if resp.Array.State != "STARTED" {
		t.Errorf("expected state STARTED, got %s", resp.Array.State)
	}
	if resp.Array.Capacity.Kilobytes.Total != "97990693745" {
		t.Errorf("expected total capacity 97990693745, got %s", resp.Array.Capacity.Kilobytes.Total)
	}
	if len(resp.Array.Disks) != 3 {
		t.Errorf("expected 3 disks, got %d", len(resp.Array.Disks))
	}

	// Check first disk
	disk := resp.Array.Disks[0]
	if disk.Name != "disk1" {
		t.Errorf("expected disk name disk1, got %s", disk.Name)
	}
	if disk.Status != "DISK_OK" {
		t.Errorf("expected status DISK_OK, got %s", disk.Status)
	}
	if disk.Temp != 34 {
		t.Errorf("expected temp 34, got %d", disk.Temp)
	}
}

// TestIntegration_DockerContainersQuery tests the docker containers query against the mock Unraid API
func TestIntegration_DockerContainersQuery(t *testing.T) {
	server := MockUnraidAPI()
	defer server.Close()

	c := client.New(server.URL, "test-api-key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp DockerResponse
	err := c.Query(ctx, DockerContainersQuery, nil, &resp)
	if err != nil {
		t.Fatalf("DockerContainersQuery failed: %v", err)
	}

	// Verify response
	containers := resp.Docker.Containers
	if len(containers) != 3 {
		t.Fatalf("expected 3 containers, got %d", len(containers))
	}

	// Check plex container
	plex := containers[0]
	if plex.Names[0] != "/plex" {
		t.Errorf("expected name /plex, got %s", plex.Names[0])
	}
	if plex.State != "RUNNING" {
		t.Errorf("expected state RUNNING, got %s", plex.State)
	}
	if !plex.AutoStart {
		t.Error("expected autoStart true for plex")
	}

	// Check exited container
	radarr := containers[2]
	if radarr.State != "EXITED" {
		t.Errorf("expected state EXITED, got %s", radarr.State)
	}
	if radarr.AutoStart {
		t.Error("expected autoStart false for radarr")
	}
}

// TestIntegration_SharesQuery tests the shares query against the mock Unraid API
func TestIntegration_SharesQuery(t *testing.T) {
	server := MockUnraidAPI()
	defer server.Close()

	c := client.New(server.URL, "test-api-key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp SharesResponse
	err := c.Query(ctx, SharesQuery, nil, &resp)
	if err != nil {
		t.Fatalf("SharesQuery failed: %v", err)
	}

	// Verify response
	if len(resp.Shares) != 3 {
		t.Fatalf("expected 3 shares, got %d", len(resp.Shares))
	}

	// Check appdata share
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

// TestIntegration_NotificationsQuery tests the notifications query against the mock Unraid API
func TestIntegration_NotificationsQuery(t *testing.T) {
	server := MockUnraidAPI()
	defer server.Close()

	c := client.New(server.URL, "test-api-key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp NotificationsResponse
	err := c.Query(ctx, NotificationsQuery, nil, &resp)
	if err != nil {
		t.Fatalf("NotificationsQuery failed: %v", err)
	}

	// Verify response
	if resp.Notifications.Overview.Unread.Total != 3 {
		t.Errorf("expected 3 unread, got %d", resp.Notifications.Overview.Unread.Total)
	}
	if len(resp.Notifications.List) != 3 {
		t.Fatalf("expected 3 notifications, got %d", len(resp.Notifications.List))
	}

	// Check first notification
	notif := resp.Notifications.List[0]
	if notif.Subject != "Parity check complete" {
		t.Errorf("expected subject 'Parity check complete', got %s", notif.Subject)
	}
	if notif.Importance != "WARNING" {
		t.Errorf("expected importance WARNING, got %s", notif.Importance)
	}
}

// TestIntegration_VMsQuery tests the VMs query against the mock Unraid API
func TestIntegration_VMsQuery(t *testing.T) {
	server := MockUnraidAPI()
	defer server.Close()

	c := client.New(server.URL, "test-api-key")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp VMsResponse
	err := c.Query(ctx, VMsQuery, nil, &resp)
	if err != nil {
		t.Fatalf("VMsQuery failed: %v", err)
	}

	// Verify response
	if len(resp.VMs.Domain) != 2 {
		t.Fatalf("expected 2 VMs, got %d", len(resp.VMs.Domain))
	}

	// Check Windows VM
	win := resp.VMs.Domain[0]
	if win.Name != "Windows 11" {
		t.Errorf("expected name 'Windows 11', got %s", win.Name)
	}
	if win.State != "running" {
		t.Errorf("expected state running, got %s", win.State)
	}

	// Check Ubuntu VM
	ubuntu := resp.VMs.Domain[1]
	if ubuntu.State != "shutoff" {
		t.Errorf("expected state shutoff, got %s", ubuntu.State)
	}
}

// TestIntegration_Unauthorized tests that requests without API key are rejected
func TestIntegration_Unauthorized(t *testing.T) {
	server := MockUnraidAPI()
	defer server.Close()

	c := client.New(server.URL, "") // No API key
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp InfoResponse
	err := c.Query(ctx, InfoQuery, nil, &resp)
	if err == nil {
		t.Fatal("expected error for unauthorized request, got nil")
	}
}
