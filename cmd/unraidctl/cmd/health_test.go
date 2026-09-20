package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/jwmoss/unraidctl/pkg/client"
)

func TestUPSStatusUsesAlarmFlags(t *testing.T) {
	for _, tc := range []struct{ status, want string }{
		{"ONLINE REPLACEBATT", "ALERT"}, {"ONLINE LOWBATT", "ALERT"}, {"ONLINE COMMLOST", "UNKNOWN"},
		{"ONBATT", "WARNING"}, {"ONLINE", "OK"}, {"", "UNKNOWN"}, {"unexpected", "UNKNOWN"},
	} {
		t.Run(tc.status, func(t *testing.T) {
			if got := upsStatusLevel(tc.status); got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
	var report healthReport
	if err := json.Unmarshal([]byte(`{"upsDevices":[{"name":"UPS","status":"ONLINE REPLACEBATT","battery":{"health":"Good"}}]}`), &report); err != nil {
		t.Fatal(err)
	}
	checks := assessHealth(report)
	for _, check := range checks {
		if check.Component == "UPS: UPS" && check.Status == "ALERT" {
			return
		}
	}
	t.Fatal("upstream Good value hid the replace-battery alarm")
}

func TestUnknownParityErrorsStayUnknown(t *testing.T) {
	check := parityHealthCheck([]api.ParityCheck{{Status: "COMPLETED"}})
	if check.Status != "UNKNOWN" || !strings.Contains(check.Message, "unknown") {
		t.Fatalf("unknown errors became healthy: %+v", check)
	}
}

func TestHealthPreservesResultsOnDeniedUPS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req client.GraphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		response := `{"data":{}}`
		switch req.Query {
		case api.UPSQuery:
			response = `{"errors":[{"message":"Forbidden"}]}`
		case api.ArrayStatusQuery:
			response = `{"data":{"array":{"state":"STARTED","disks":[{"name":"disk1","device":"sda","size":1,"status":"DISK_OK"}]}}}`
		case api.NotificationAlertsQuery:
			response = `{"data":{"notifications":{"overview":{"unread":{"alert":2}}}}}`
		}
		_, _ = io.WriteString(w, response)
	}))
	defer server.Close()
	old := apiClient
	apiClient = client.New(server.URL, "test-key")
	defer func() { apiClient = old }()
	report := readHealth()
	if report.Status != "INCOMPLETE" || len(report.Errors) != 1 {
		t.Fatalf("expected incomplete report: %+v", report)
	}
	if report.Array.State != "STARTED" || report.Notifications.Overview.Unread.Alert != 2 {
		t.Fatalf("lost successful checks: %+v", report)
	}
}

func TestArrayChecksIgnoreUnassignedSlotsButIncludeCacheErrors(t *testing.T) {
	errors := int64(7)
	checks := arrayHealthChecks(api.Array{State: "STARTED", Parities: []api.ArrayDisk{{Name: "parity2", Status: "DISK_NP_DSBL"}}, Caches: []api.ArrayDisk{{Name: "cache", Device: "sdf", Status: "DISK_OK", NumErrors: &errors}}})
	if len(checks) != 2 || checks[1].Component != "Disk: cache" || checks[1].Status != "ALERT" {
		t.Fatalf("unexpected disk checks: %+v", checks)
	}
}
