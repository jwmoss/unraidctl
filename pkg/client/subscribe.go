package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// graphqlWSProtocol is the graphql-ws subprotocol that the Unraid API serves on /graphql.
const graphqlWSProtocol = "graphql-transport-ws"

// ErrStopSubscription ends a subscription without an error when a handler returns it.
var ErrStopSubscription = errors.New("stop subscription")

type wsMessage struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Subscribe runs a GraphQL subscription and calls handle with the data of each event.
// It returns nil when the server completes, ctx ends, or handle returns ErrStopSubscription.
// The client timeout applies to the handshake only.
func (c *Client) Subscribe(ctx context.Context, query string, variables map[string]interface{}, handle func(json.RawMessage) error) error {
	document, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		return fmt.Errorf("cannot parse GraphQL subscription: %w", err)
	}
	for _, operation := range document.Operations {
		if operation.Operation != ast.Subscription {
			return fmt.Errorf("refusing GraphQL %s over a subscription", operation.Operation)
		}
	}

	conn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.CloseNow() }()

	if err := writeMessage(ctx, conn, wsMessage{ID: "1", Type: "subscribe", Payload: mustJSON(GraphQLRequest{Query: query, Variables: variables})}); err != nil {
		return err
	}
	for {
		message, err := readMessage(ctx, conn)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("subscription failed: %w", err)
		}
		if err := c.dispatch(ctx, conn, message, handle); err != nil {
			_ = writeMessage(ctx, conn, wsMessage{ID: "1", Type: "complete"})
			_ = conn.Close(websocket.StatusNormalClosure, "")
			if errors.Is(err, ErrStopSubscription) {
				return nil
			}
			return err
		}
	}
}

// dispatch handles one server message; it returns ErrStopSubscription when the server completes.
func (c *Client) dispatch(ctx context.Context, conn *websocket.Conn, message wsMessage, handle func(json.RawMessage) error) error {
	switch message.Type {
	case "next":
		var payload GraphQLResponse
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			return fmt.Errorf("failed to parse subscription event: %w", err)
		}
		if len(payload.Errors) > 0 {
			return fmt.Errorf("GraphQL error: %s", c.redact(payload.Errors[0].Message))
		}
		return handle(payload.Data)
	case "error":
		var errs []GraphQLError
		if json.Unmarshal(message.Payload, &errs) != nil || len(errs) == 0 {
			return fmt.Errorf("GraphQL subscription error: %s", c.redact(string(message.Payload)))
		}
		return fmt.Errorf("GraphQL error: %s", c.redact(errs[0].Message))
	case "complete":
		return ErrStopSubscription
	case "ping":
		return writeMessage(ctx, conn, wsMessage{Type: "pong"})
	}
	return nil
}

// connect opens the WebSocket and completes the graphql-ws connection_init handshake.
func (c *Client) connect(ctx context.Context) (*websocket.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, c.httpClient.Timeout)
	defer cancel()

	endpoint := c.baseURL + "/graphql"
	if !strings.HasPrefix(strings.ToLower(endpoint), "http") {
		return nil, fmt.Errorf("server URL must use http or https")
	}
	// The websocket library rejects clients with a timeout; the context bounds the handshake.
	httpClient := *c.httpClient
	httpClient.Timeout = 0
	start := time.Now()
	conn, resp, err := websocket.Dial(ctx, "ws"+endpoint[len("http"):], &websocket.DialOptions{
		HTTPClient:   &httpClient,
		HTTPHeader:   http.Header{"x-api-key": []string{c.apiKey}},
		Subprotocols: []string{graphqlWSProtocol},
	})
	if c.trace != nil && resp != nil {
		c.trace(http.MethodGet, "/graphql", resp.StatusCode, time.Since(start))
	}
	if err != nil {
		return nil, fmt.Errorf("subscription connection failed: %s", c.redact(err.Error()))
	}
	conn.SetReadLimit(64 << 20)

	err = writeMessage(ctx, conn, wsMessage{Type: "connection_init", Payload: mustJSON(map[string]string{"x-api-key": c.apiKey})})
	for err == nil {
		var message wsMessage
		if message, err = readMessage(ctx, conn); err == nil && message.Type == "connection_ack" {
			return conn, nil
		}
	}
	_ = conn.CloseNow()
	return nil, fmt.Errorf("subscription handshake failed: %w", err)
}

func writeMessage(ctx context.Context, conn *websocket.Conn, message wsMessage) error {
	if err := conn.Write(ctx, websocket.MessageText, mustJSON(message)); err != nil {
		return fmt.Errorf("failed to send %s message: %w", message.Type, err)
	}
	return nil
}

func readMessage(ctx context.Context, conn *websocket.Conn) (wsMessage, error) {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return wsMessage{}, err
	}
	var message wsMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return wsMessage{}, fmt.Errorf("invalid subscription message: %w", err)
	}
	return message, nil
}

func mustJSON(value interface{}) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return data
}
