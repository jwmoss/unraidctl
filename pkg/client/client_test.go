package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_Query_Success(t *testing.T) {
	// Mock server that returns a successful GraphQL response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("expected x-api-key test-api-key, got %s", r.Header.Get("x-api-key"))
		}

		// Return mock response
		response := GraphQLResponse{
			Data: json.RawMessage(`{"info":{"os":{"hostname":"TestServer"}}}`),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := New(server.URL, "test-api-key")
	ctx := context.Background()

	var result map[string]interface{}
	err := client.Query(ctx, "{ info { os { hostname } } }", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, ok := result["info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected info in result")
	}
	os, ok := info["os"].(map[string]interface{})
	if !ok {
		t.Fatal("expected os in info")
	}
	hostname, ok := os["hostname"].(string)
	if !ok || hostname != "TestServer" {
		t.Errorf("expected hostname TestServer, got %v", os["hostname"])
	}
}

func TestClient_Query_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GraphQLResponse{
			Errors: []GraphQLError{
				{Message: "Field not found"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := New(server.URL, "test-api-key")
	ctx := context.Background()

	var result map[string]interface{}
	err := client.Query(ctx, "{ invalid }", nil, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "GraphQL error: Field not found" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestClient_Query_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte("Unauthorized")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := New(server.URL, "bad-api-key")
	ctx := context.Background()

	var result map[string]interface{}
	err := client.Query(ctx, "{ info }", nil, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Query_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, "test-api-key")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var result map[string]interface{}
	err := client.Query(ctx, "{ info }", nil, &result)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestClient_Query_WithVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Verify variables were sent
		if req.Variables == nil {
			t.Fatal("expected variables in request")
		}
		if req.Variables["id"] != "container-123" {
			t.Errorf("expected id container-123, got %v", req.Variables["id"])
		}

		response := GraphQLResponse{
			Data: json.RawMessage(`{"container":{"id":"container-123","state":"running"}}`),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := New(server.URL, "test-api-key")
	ctx := context.Background()

	vars := map[string]interface{}{"id": "container-123"}
	var result map[string]interface{}
	err := client.Query(ctx, "query($id: String!) { container(id: $id) { id state } }", vars, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// roundTripFunc replaces only the HTTP wire, so Client.Query still uses the
// production HTTP client and its normal redirect callback.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClient_Query_RedirectOrigins(t *testing.T) {
	cases := []struct {
		name, from, to, wantError string
		wantRequests              int
		loop                      bool
	}{
		{"implicit HTTP port", "http://fixture.invalid", "HTTP://FIXTURE.INVALID:80/next", "", 2, false},
		{"zero-padded HTTP port", "http://fixture.invalid", "http://fixture.invalid:00080/next", "", 2, false},
		{"explicit HTTP port", "http://fixture.invalid:80", "http://fixture.invalid/next", "", 2, false},
		{"implicit HTTPS port", "https://fixture.invalid", "https://fixture.invalid:443/next", "", 2, false},
		{"explicit HTTPS port", "https://fixture.invalid:443", "https://FIXTURE.INVALID/next", "", 2, false},
		{"default to zero port", "http://fixture.invalid", "http://fixture.invalid:0/next", "cross-origin", 1, false},
		{"zero to default port", "http://fixture.invalid:0", "http://fixture.invalid/next", "cross-origin", 1, false},
		{"different port", "http://fixture.invalid:8080", "http://fixture.invalid:80/next", "cross-origin", 1, false},
		{"different host", "http://fixture.invalid", "http://other.invalid/next", "cross-origin", 1, false},
		{"different scheme", "http://fixture.invalid:80", "https://fixture.invalid:80/next", "cross-origin", 1, false},
		{"downgrade", "https://fixture.invalid:443", "http://fixture.invalid:443/next", "cross-origin", 1, false},
		{"redirect limit", "http://fixture.invalid", "http://fixture.invalid/next", "stopped after 10 redirects", 10, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := New(tc.from, "synthetic-redirect-key")
			requests := 0
			c.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				requests++
				if req.Method != http.MethodPost || req.Header.Get("x-api-key") != "synthetic-redirect-key" {
					t.Errorf("redirect must preserve POST and the API key")
				}
				if req.Body != nil {
					_ = req.Body.Close()
				}
				response := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":{"marker":"redirect-success"}}`)), Request: req}
				if requests == 1 || tc.loop {
					response.StatusCode = http.StatusTemporaryRedirect
					response.Header.Set("Location", tc.to)
				}
				return response, nil
			})
			var result map[string]string
			err := c.Query(context.Background(), "query { marker }", nil, &result)
			if tc.wantError == "" {
				if err != nil || result["marker"] != "redirect-success" {
					t.Fatalf("same-origin query fails: result=%v error=%v", result, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("expected %q, got %v", tc.wantError, err)
			}
			if requests != tc.wantRequests {
				t.Errorf("sent %d requests; want %d", requests, tc.wantRequests)
			}
		})
	}
}
