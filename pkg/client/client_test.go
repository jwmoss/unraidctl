package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		w.Write([]byte("Unauthorized"))
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
		json.NewEncoder(w).Encode(response)
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
