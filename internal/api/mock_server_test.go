package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockUnraidAPI creates a mock HTTP server that simulates the Unraid GraphQL API
// Note: The client appends /graphql to the base URL, so we need to handle that path
func MockUnraidAPI() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", handleGraphQL)
	return httptest.NewServer(mux)
}

func handleGraphQL(w http.ResponseWriter, r *http.Request) {
	// Verify it's a POST request
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify API key
	apiKey := r.Header.Get("x-api-key")
	if apiKey == "" {
		http.Error(w, `{"errors":[{"message":"Unauthorized"}]}`, http.StatusUnauthorized)
		return
	}

	// Parse the GraphQL query
	var req struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"errors":[{"message":"Invalid request"}]}`, http.StatusBadRequest)
		return
	}

	// Route based on query content
	w.Header().Set("Content-Type", "application/json")

	var response interface{}
	switch {
	case strings.Contains(req.Query, "metrics") && strings.Contains(req.Query, "network"):
		response = mockNetworkMetricsResponse()
	case strings.Contains(req.Query, "restart"):
		response = mockDockerRestartResponse()
	case strings.Contains(req.Query, "apiKeyPossibleRoles"):
		response = mockAPIKeyMetadataResponse()
	case strings.Contains(req.Query, "apiKeys"):
		response = mockAPIKeysResponse()
	case strings.Contains(req.Query, "apiKey") && strings.Contains(req.Query, "create"):
		response = mockCreateAPIKeyResponse()
	case strings.Contains(req.Query, "setState"):
		response = mockArraySetStateResponse()
	case strings.Contains(req.Query, "logs("):
		response = mockDockerLogsResponse()
	case strings.Contains(req.Query, "logFiles"):
		response = mockLogFilesResponse()
	case strings.Contains(req.Query, "logFile("):
		response = mockLogFileResponse()
	case strings.Contains(req.Query, "settings"):
		response = mockSettingsResponse()
	case strings.Contains(req.Query, "validateOidcSession"):
		response = mockValidateOIDCSessionResponse()
	case strings.Contains(req.Query, "oidcConfiguration"):
		response = mockOIDCConfigurationResponse()
	case strings.Contains(req.Query, "oidcProviders") || strings.Contains(req.Query, "publicOidcProviders"):
		response = mockOIDCProvidersResponse()
	case strings.Contains(req.Query, "info"):
		response = mockInfoResponse()
	case strings.Contains(req.Query, "array"):
		response = mockArrayResponse()
	case strings.Contains(req.Query, "docker"):
		response = mockDockerResponse()
	case strings.Contains(req.Query, "shares"):
		response = mockSharesResponse()
	case strings.Contains(req.Query, "notifications"):
		response = mockNotificationsResponse()
	case strings.Contains(req.Query, "vms"):
		response = mockVMsResponse()
	default:
		response = map[string]interface{}{
			"errors": []map[string]string{
				{"message": "Unknown query"},
			},
		}
	}
	_ = json.NewEncoder(w).Encode(response)
}

func mockNetworkMetricsResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"metrics": map[string]interface{}{
				"network": []map[string]interface{}{
					{
						"id":                 "metrics/network/eth0",
						"name":               "eth0",
						"operstate":          "up",
						"bytesReceived":      1024,
						"bytesSent":          2048,
						"packetsReceived":    10,
						"packetsSent":        20,
						"receiveErrors":      1,
						"transmitErrors":     2,
						"receiveDropped":     3,
						"transmitDropped":    4,
						"rxSec":              100.5,
						"txSec":              200.5,
						"utilizationPercent": 0.0024,
						"lastUpdated":        "2026-08-19T12:00:00.000Z",
					},
				},
			},
		},
	}
}

func mockDockerRestartResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"docker": map[string]interface{}{
				"restart": map[string]interface{}{
					"id":        "abc123def456",
					"names":     []string{"/plex"},
					"state":     "RUNNING",
					"status":    "Up 1 second",
					"image":     "linuxserver/plex:latest",
					"autoStart": true,
				},
			},
		},
	}
}

func mockInfoResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"info": map[string]interface{}{
				"os": map[string]interface{}{
					"platform": "linux",
					"distro":   "Unraid OS",
					"release":  "7.2 x86_64",
					"uptime":   "2026-01-16T15:07:42.840Z",
					"hostname": "Tower",
				},
				"cpu": map[string]interface{}{
					"manufacturer": "Intel",
					"brand":        "Xeon E3-1246 v3",
					"cores":        4,
					"speed":        3.5,
				},
			},
		},
	}
}

func mockAPIKeysResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"apiKeys": []map[string]interface{}{
				{
					"id":          "key-1",
					"name":        "automation",
					"description": "automation key",
					"roles":       []string{"VIEWER"},
					"createdAt":   "2026-05-01T12:00:00.000Z",
					"permissions": []map[string]interface{}{
						{"resource": "INFO", "actions": []string{"READ_ANY"}},
					},
				},
			},
		},
	}
}

func mockAPIKeyMetadataResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"apiKeyPossibleRoles": []string{"ADMIN", "VIEWER", "GUEST"},
			"apiKeyPossiblePermissions": []map[string]interface{}{
				{"resource": "INFO", "actions": []string{"READ_ANY"}},
				{"resource": "DOCKER", "actions": []string{"READ_ANY", "UPDATE_ANY"}},
			},
			"getAvailableAuthActions": []string{"READ_ANY", "UPDATE_ANY"},
		},
	}
}

func mockCreateAPIKeyResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"apiKey": map[string]interface{}{
				"create": map[string]interface{}{
					"id":          "key-2",
					"key":         "secret-value",
					"name":        "new-key",
					"description": "new key",
					"roles":       []string{"VIEWER"},
					"createdAt":   "2026-05-02T12:00:00.000Z",
					"permissions": []map[string]interface{}{},
				},
			},
		},
	}
}

func mockArraySetStateResponse() map[string]interface{} {
	resp := mockArrayResponse()
	array := resp["data"].(map[string]interface{})["array"]
	return map[string]interface{}{
		"data": map[string]interface{}{
			"array": map[string]interface{}{
				"setState": array,
			},
		},
	}
}

func mockArrayResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"array": map[string]interface{}{
				"state": "STARTED",
				"capacity": map[string]interface{}{
					"kilobytes": map[string]interface{}{
						"free":  "29592689497",
						"used":  "68398004248",
						"total": "97990693745",
					},
				},
				"disks": []map[string]interface{}{
					{
						"id":     "disk1-id",
						"name":   "disk1",
						"device": "sdd",
						"size":   13672382412,
						"status": "DISK_OK",
						"temp":   34,
						"type":   "DATA",
					},
					{
						"id":     "disk2-id",
						"name":   "disk2",
						"device": "sde",
						"size":   13672382412,
						"status": "DISK_OK",
						"temp":   36,
						"type":   "DATA",
					},
					{
						"id":     "parity-id",
						"name":   "parity",
						"device": "sda",
						"size":   13672382412,
						"status": "DISK_OK",
						"temp":   32,
						"type":   "PARITY",
					},
				},
			},
		},
	}
}

func mockDockerResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"docker": map[string]interface{}{
				"containers": []map[string]interface{}{
					{
						"id":        "abc123def456",
						"names":     []string{"/plex"},
						"state":     "RUNNING",
						"status":    "Up 16 hours",
						"image":     "linuxserver/plex:latest",
						"autoStart": true,
					},
					{
						"id":        "def456ghi789",
						"names":     []string{"/sonarr"},
						"state":     "RUNNING",
						"status":    "Up 18 hours",
						"image":     "linuxserver/sonarr:develop",
						"autoStart": true,
					},
					{
						"id":        "ghi789jkl012",
						"names":     []string{"/radarr"},
						"state":     "EXITED",
						"status":    "Exited (0) 2 hours ago",
						"image":     "linuxserver/radarr:latest",
						"autoStart": false,
					},
				},
			},
		},
	}
}

func mockDockerLogsResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"docker": map[string]interface{}{
				"logs": map[string]interface{}{
					"containerId": "abc123def456",
					"cursor":      "2026-05-29T12:00:01.000Z",
					"lines": []map[string]interface{}{
						{
							"timestamp": "2026-05-29T12:00:00.000Z",
							"message":   "server started",
						},
						{
							"timestamp": "2026-05-29T12:00:01.000Z",
							"message":   "ready",
						},
					},
				},
			},
		},
	}
}

func mockSharesResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"shares": []map[string]interface{}{
				{
					"name":    "appdata",
					"comment": "application data",
					"free":    938371584,
					"used":    83130216,
				},
				{
					"name":    "media",
					"comment": "movies and tv shows",
					"free":    29592689496,
					"used":    68398004249,
				},
				{
					"name":    "downloads",
					"comment": "",
					"free":    29592689496,
					"used":    1234567890,
				},
			},
		},
	}
}

func mockNotificationsResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"notifications": map[string]interface{}{
				"overview": map[string]interface{}{
					"unread": map[string]interface{}{
						"total": 3,
					},
				},
				"list": []map[string]interface{}{
					{
						"id":         "notif-1",
						"subject":    "Parity check complete",
						"importance": "WARNING",
						"timestamp":  "2026-01-16T15:15:01.000Z",
					},
					{
						"id":         "notif-2",
						"subject":    "Docker update available",
						"importance": "INFO",
						"timestamp":  "2026-01-17T10:00:00.000Z",
					},
					{
						"id":         "notif-3",
						"subject":    "Array started",
						"importance": "INFO",
						"timestamp":  "2026-01-18T08:30:00.000Z",
					},
				},
			},
		},
	}
}

func mockVMsResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"vms": map[string]interface{}{
				"id": "vms-node-id",
				"domain": []map[string]interface{}{
					{
						"name":  "Windows 11",
						"state": "running",
					},
					{
						"name":  "Ubuntu Server",
						"state": "shutoff",
					},
				},
			},
		},
	}
}

func mockLogFilesResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"logFiles": []map[string]interface{}{
				{
					"name":       "graphql-api.log",
					"path":       "/var/log/graphql-api.log",
					"size":       2048,
					"modifiedAt": "2026-05-29T12:00:00.000Z",
				},
			},
		},
	}
}

func mockLogFileResponse() map[string]interface{} {
	startLine := 1
	return map[string]interface{}{
		"data": map[string]interface{}{
			"logFile": map[string]interface{}{
				"path":       "/var/log/graphql-api.log",
				"content":    "line 1\nline 2\n",
				"totalLines": 2,
				"startLine":  startLine,
			},
		},
	}
}

func mockSettingsResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"isSSOEnabled": true,
			"settings": map[string]interface{}{
				"id": "settings-id",
				"api": map[string]interface{}{
					"version":      "4.0.0",
					"extraOrigins": []string{"https://example.test"},
					"sandbox":      true,
					"ssoSubIds":    []string{"user-1"},
					"plugins":      []string{"connect"},
				},
				"sso": map[string]interface{}{
					"id": "sso-id",
					"oidcProviders": []map[string]interface{}{
						{
							"id":            "oidc-1",
							"name":          "Google",
							"clientId":      "client-id",
							"issuer":        "https://accounts.google.com",
							"scopes":        []string{"openid", "email"},
							"buttonText":    "Sign in",
							"buttonVariant": "default",
						},
					},
				},
				"unified": map[string]interface{}{
					"values": map[string]interface{}{"api": map[string]interface{}{"sandbox": true}},
				},
			},
		},
	}
}

func mockOIDCProvidersResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"isSSOEnabled": true,
			"oidcProviders": []map[string]interface{}{
				{
					"id":            "oidc-1",
					"name":          "Google",
					"clientId":      "client-id",
					"issuer":        "https://accounts.google.com",
					"scopes":        []string{"openid", "email"},
					"buttonText":    "Sign in",
					"buttonVariant": "default",
				},
			},
			"publicOidcProviders": []map[string]interface{}{
				{
					"id":            "public-1",
					"name":          "Google",
					"buttonText":    "Sign in",
					"buttonIcon":    "",
					"buttonVariant": "default",
					"buttonStyle":   "",
				},
			},
		},
	}
}

func mockOIDCConfigurationResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"oidcConfiguration": map[string]interface{}{
				"defaultAllowedOrigins": []string{"https://tower.local"},
				"providers": []map[string]interface{}{
					{
						"id":            "oidc-1",
						"name":          "Google",
						"clientId":      "client-id",
						"issuer":        "https://accounts.google.com",
						"scopes":        []string{"openid", "email"},
						"buttonText":    "Sign in",
						"buttonVariant": "default",
					},
				},
			},
		},
	}
}

func mockValidateOIDCSessionResponse() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"validateOidcSession": map[string]interface{}{
				"valid":    true,
				"username": "jonathan",
			},
		},
	}
}
