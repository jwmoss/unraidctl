package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockUnraidAPI creates a mock HTTP server that simulates the Unraid GraphQL API
func MockUnraidAPI() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a GraphQL request
		if r.Method != "POST" || r.URL.Path != "/graphql" {
			http.Error(w, "Not Found", http.StatusNotFound)
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

		switch {
		case strings.Contains(req.Query, "info"):
			json.NewEncoder(w).Encode(mockInfoResponse())
		case strings.Contains(req.Query, "array"):
			json.NewEncoder(w).Encode(mockArrayResponse())
		case strings.Contains(req.Query, "docker"):
			json.NewEncoder(w).Encode(mockDockerResponse())
		case strings.Contains(req.Query, "shares"):
			json.NewEncoder(w).Encode(mockSharesResponse())
		case strings.Contains(req.Query, "notifications"):
			json.NewEncoder(w).Encode(mockNotificationsResponse())
		case strings.Contains(req.Query, "vms"):
			json.NewEncoder(w).Encode(mockVMsResponse())
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]string{
					{"message": "Unknown query"},
				},
			})
		}
	}))
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
