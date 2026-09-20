package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/jwmoss/unraidctl/pkg/client"
)

func TestNotificationPagination(t *testing.T) {
	requests := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req client.GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		filter := req.Variables["filter"].(map[string]interface{})
		kind := filter["type"].(string)
		offset := int(filter["offset"].(float64))
		requests = append(requests, fmt.Sprintf("%s:%d", kind, offset))
		count := 101
		if kind == "ARCHIVE" {
			count = 2
		}
		notes := make([]api.Notification, 0)
		for i := offset; i < count && i < offset+100; i++ {
			notes = append(notes, api.Notification{ID: fmt.Sprintf("%s-%d", kind, i), Type: kind, Timestamp: "2026-09-20T10:00:00Z"})
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"notifications": map[string]interface{}{"list": notes}}})
	}))
	defer server.Close()
	old := apiClient
	apiClient = client.New(server.URL, "test-key")
	defer func() { apiClient = old }()
	notes, err := listNotifications(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 103 {
		t.Fatalf("got %d notifications", len(notes))
	}
	if got := strings.Join(requests, ","); got != "UNREAD:0,UNREAD:100,ARCHIVE:0" {
		t.Fatalf("unexpected pages: %s", got)
	}
}

func TestNotificationPaginationFailureIsNotPartialSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req client.GraphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		filter := req.Variables["filter"].(map[string]interface{})
		if filter["offset"].(float64) > 0 {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"notifications": map[string]interface{}{"list": make([]api.Notification, 100)}}})
	}))
	defer server.Close()
	old := apiClient
	apiClient = client.New(server.URL, "test-key")
	defer func() { apiClient = old }()
	notes, err := listNotifications(false)
	if err == nil || notes != nil {
		t.Fatalf("partial result reported as success: %d %v", len(notes), err)
	}
}
