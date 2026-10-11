package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	dryRun     bool
	trace      func(string, string, int, time.Duration)
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
	Path    []any  `json:"path,omitempty"`
}

type Option func(*Client)

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = timeout }
}

func WithTrace(trace func(string, string, int, time.Duration)) Option {
	return func(c *Client) { c.trace = trace }
}

func WithDryRun(enabled bool) Option {
	return func(c *Client) { c.dryRun = enabled }
}

func New(baseURL, apiKey string, options ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// net/http's transport requires a lowercase scheme, including redirects.
				req.URL.Scheme = strings.ToLower(req.URL.Scheme)
				original := via[0].URL
				if !strings.EqualFold(req.URL.Scheme, original.Scheme) ||
					!strings.EqualFold(req.URL.Hostname(), original.Hostname()) ||
					effectivePort(req.URL) != effectivePort(original) {
					return fmt.Errorf("refusing cross-origin API redirect")
				}
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	if c.dryRun {
		document, err := parser.ParseQuery(&ast.Source{Input: query})
		if err != nil {
			return fmt.Errorf("dry-run: cannot validate GraphQL operation: %w", err)
		}
		for _, operation := range document.Operations {
			if operation.Operation != ast.Query {
				return fmt.Errorf("dry-run: refusing GraphQL %s", operation.Operation)
			}
		}
	}
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/graphql", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if c.trace != nil {
		c.trace(req.Method, c.redact(req.URL.Path), resp.StatusCode, time.Since(start))
	}

	const maxResponseBytes = 64 << 20
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if len(respBody) > maxResponseBytes {
		return fmt.Errorf("response exceeds the 64 MiB limit")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, c.redact(string(respBody)))
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", c.redact(gqlResp.Errors[0].Message))
	}

	if result != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

func (c *Client) redact(message string) string {
	if c.apiKey == "" {
		return message
	}
	return strings.ReplaceAll(message, c.apiKey, "[REDACTED]")
}

func effectivePort(endpoint *url.URL) string {
	if port := endpoint.Port(); port != "" {
		if normalized := strings.TrimLeft(port, "0"); normalized != "" {
			return normalized
		}
		return "0"
	}
	switch strings.ToLower(endpoint.Scheme) {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}
