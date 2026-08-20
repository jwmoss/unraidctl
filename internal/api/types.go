package api

import "encoding/json"

// Response types for Unraid API
// Based on actual schema from https://docs.unraid.net/API/

type InfoResponse struct {
	Info struct {
		OS struct {
			Platform string `json:"platform"`
			Distro   string `json:"distro"`
			Release  string `json:"release"`
			Uptime   string `json:"uptime"` // ISO 8601 timestamp of boot time
			Hostname string `json:"hostname"`
		} `json:"os"`
		CPU struct {
			Manufacturer string  `json:"manufacturer"`
			Brand        string  `json:"brand"`
			Cores        int     `json:"cores"`
			Speed        float64 `json:"speed"`
		} `json:"cpu"`
	} `json:"info"`
}

type ArrayResponse struct {
	Array Array `json:"array"`
}

type Array struct {
	State    string `json:"state"`
	Capacity struct {
		Kilobytes struct {
			Free  string `json:"free"` // String in API response
			Used  string `json:"used"`
			Total string `json:"total"`
		} `json:"kilobytes"`
	} `json:"capacity"`
	Disks []ArrayDisk `json:"disks"`
}

type ArrayMutationResponse struct {
	Array struct {
		SetState                 Array     `json:"setState"`
		AddDiskToArray           Array     `json:"addDiskToArray"`
		RemoveDiskFromArray      Array     `json:"removeDiskFromArray"`
		MountArrayDisk           ArrayDisk `json:"mountArrayDisk"`
		UnmountArrayDisk         ArrayDisk `json:"unmountArrayDisk"`
		ClearArrayDiskStatistics bool      `json:"clearArrayDiskStatistics"`
	} `json:"array"`
}

type ArrayDisk struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Device string `json:"device"`
	Size   int64  `json:"size"` // in KB
	Status string `json:"status"`
	Temp   int    `json:"temp"`
	Type   string `json:"type"`
}

type DockerResponse struct {
	Docker struct {
		Containers []DockerContainer `json:"containers"`
	} `json:"docker"`
}

type DockerContainer struct {
	ID                string               `json:"id"`
	Names             []string             `json:"names"`
	State             string               `json:"state"`
	Status            string               `json:"status"`
	Image             string               `json:"image"`
	ImageID           string               `json:"imageId,omitempty"`
	Command           string               `json:"command,omitempty"`
	Created           int64                `json:"created,omitempty"`
	LanIPPorts        []string             `json:"lanIpPorts,omitempty"`
	SizeRootFs        *int64               `json:"sizeRootFs,omitempty"`
	SizeRw            *int64               `json:"sizeRw,omitempty"`
	SizeLog           *int64               `json:"sizeLog,omitempty"`
	AutoStart         bool                 `json:"autoStart"`
	AutoStartOrder    *int                 `json:"autoStartOrder,omitempty"`
	AutoStartWait     *int                 `json:"autoStartWait,omitempty"`
	Ports             []ContainerPort      `json:"ports,omitempty"`
	HostConfig        *ContainerHostConfig `json:"hostConfig,omitempty"`
	NetworkSettings   json.RawMessage      `json:"networkSettings,omitempty"`
	Mounts            []json.RawMessage    `json:"mounts,omitempty"`
	IsOrphaned        bool                 `json:"isOrphaned,omitempty"`
	IsUpdateAvailable *bool                `json:"isUpdateAvailable,omitempty"`
	IsRebuildReady    *bool                `json:"isRebuildReady,omitempty"`
	TemplatePath      string               `json:"templatePath,omitempty"`
	ProjectURL        string               `json:"projectUrl,omitempty"`
	RegistryURL       string               `json:"registryUrl,omitempty"`
	SupportURL        string               `json:"supportUrl,omitempty"`
	IconURL           string               `json:"iconUrl,omitempty"`
	WebUIURL          string               `json:"webUiUrl,omitempty"`
	Shell             string               `json:"shell,omitempty"`
	TemplatePorts     []ContainerPort      `json:"templatePorts,omitempty"`
	TailscaleEnabled  bool                 `json:"tailscaleEnabled,omitempty"`
}

type ContainerPort struct {
	PrivatePort *int   `json:"privatePort,omitempty"`
	PublicPort  *int   `json:"publicPort,omitempty"`
	Type        string `json:"type"`
}

type ContainerHostConfig struct {
	NetworkMode string `json:"networkMode"`
}

type DockerContainerResponse struct {
	Docker struct {
		Container DockerContainer `json:"container"`
	} `json:"docker"`
}

type DockerLogsResponse struct {
	Docker struct {
		Logs DockerContainerLogs `json:"logs"`
	} `json:"docker"`
}

type DockerContainerLogs struct {
	ContainerID string          `json:"containerId"`
	Cursor      string          `json:"cursor"`
	Lines       []DockerLogLine `json:"lines"`
}

type DockerLogLine struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type DockerMutationResponse struct {
	Docker struct {
		Start                        DockerContainer   `json:"start"`
		Stop                         DockerContainer   `json:"stop"`
		Restart                      DockerContainer   `json:"restart"`
		Pause                        DockerContainer   `json:"pause"`
		Unpause                      DockerContainer   `json:"unpause"`
		UpdateContainer              DockerContainer   `json:"updateContainer"`
		UpdateAllContainers          []DockerContainer `json:"updateAllContainers"`
		RemoveContainer              bool              `json:"removeContainer"`
		UpdateAutostartConfiguration bool              `json:"updateAutostartConfiguration"`
	} `json:"docker"`
}

type MetricsResponse struct {
	Metrics struct {
		Network []NetworkMetrics `json:"network"`
	} `json:"metrics"`
}

type NetworkMetrics struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Operstate          string   `json:"operstate"`
	BytesReceived      int64    `json:"bytesReceived"`
	BytesSent          int64    `json:"bytesSent"`
	PacketsReceived    int64    `json:"packetsReceived"`
	PacketsSent        int64    `json:"packetsSent"`
	ReceiveErrors      int64    `json:"receiveErrors"`
	TransmitErrors     int64    `json:"transmitErrors"`
	ReceiveDropped     int64    `json:"receiveDropped"`
	TransmitDropped    int64    `json:"transmitDropped"`
	RxSec              float64  `json:"rxSec"`
	TxSec              float64  `json:"txSec"`
	UtilizationPercent *float64 `json:"utilizationPercent"`
	LastUpdated        string   `json:"lastUpdated"`
}

type SharesResponse struct {
	Shares []Share `json:"shares"`
}

type Share struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Free    int64  `json:"free"` // in KB
	Used    int64  `json:"used"` // in KB
}

type NotificationsResponse struct {
	Notifications struct {
		Overview struct {
			Unread struct {
				Total int `json:"total"`
			} `json:"unread"`
		} `json:"overview"`
		List []Notification `json:"list"`
	} `json:"notifications"`
}

type Notification struct {
	ID         string `json:"id"`
	Subject    string `json:"subject"`
	Importance string `json:"importance"`
	Timestamp  string `json:"timestamp"`
}

type VMsResponse struct {
	VMs struct {
		ID     string `json:"id"`
		Domain []VM   `json:"domain"`
	} `json:"vms"`
}

type Permission struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

type APIKey struct {
	ID          string       `json:"id"`
	Key         string       `json:"key,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Roles       []string     `json:"roles"`
	CreatedAt   string       `json:"createdAt"`
	Permissions []Permission `json:"permissions"`
}

type APIKeysResponse struct {
	APIKeys []APIKey `json:"apiKeys"`
}

type APIKeyMetadataResponse struct {
	APIKeyPossibleRoles       []string     `json:"apiKeyPossibleRoles"`
	APIKeyPossiblePermissions []Permission `json:"apiKeyPossiblePermissions"`
	GetAvailableAuthActions   []string     `json:"getAvailableAuthActions"`
}

type APIKeyMutationResponse struct {
	APIKey struct {
		Create     APIKey `json:"create"`
		Update     APIKey `json:"update"`
		Delete     bool   `json:"delete"`
		AddRole    bool   `json:"addRole"`
		RemoveRole bool   `json:"removeRole"`
	} `json:"apiKey"`
}

type LogFilesResponse struct {
	LogFiles []LogFile `json:"logFiles"`
}

type LogFile struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt"`
}

type LogFileResponse struct {
	LogFile LogFileContent `json:"logFile"`
}

type LogFileContent struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	TotalLines int    `json:"totalLines"`
	StartLine  *int   `json:"startLine"`
}

type SettingsResponse struct {
	IsSSOEnabled bool     `json:"isSSOEnabled"`
	Settings     Settings `json:"settings"`
}

type Settings struct {
	ID      string          `json:"id"`
	API     APIConfig       `json:"api"`
	SSO     SSOSettings     `json:"sso"`
	Unified UnifiedSettings `json:"unified"`
}

type APIConfig struct {
	Version      string   `json:"version"`
	ExtraOrigins []string `json:"extraOrigins"`
	Sandbox      *bool    `json:"sandbox"`
	SSOSubIDs    []string `json:"ssoSubIds"`
	Plugins      []string `json:"plugins"`
}

type SSOSettings struct {
	ID            string         `json:"id"`
	OIDCProviders []OIDCProvider `json:"oidcProviders"`
}

type UnifiedSettings struct {
	Values json.RawMessage `json:"values"`
}

type OIDCProvider struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	ClientID              string   `json:"clientId"`
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorizationEndpoint"`
	TokenEndpoint         string   `json:"tokenEndpoint"`
	JWKSURI               string   `json:"jwksUri"`
	Scopes                []string `json:"scopes"`
	ButtonText            string   `json:"buttonText"`
	ButtonVariant         string   `json:"buttonVariant"`
}

type PublicOIDCProvider struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ButtonText    string `json:"buttonText"`
	ButtonIcon    string `json:"buttonIcon"`
	ButtonVariant string `json:"buttonVariant"`
	ButtonStyle   string `json:"buttonStyle"`
}

type OIDCProvidersResponse struct {
	IsSSOEnabled        bool                 `json:"isSSOEnabled"`
	OIDCProviders       []OIDCProvider       `json:"oidcProviders"`
	PublicOIDCProviders []PublicOIDCProvider `json:"publicOidcProviders"`
}

type OIDCConfigurationResponse struct {
	OIDCConfiguration struct {
		DefaultAllowedOrigins []string       `json:"defaultAllowedOrigins"`
		Providers             []OIDCProvider `json:"providers"`
	} `json:"oidcConfiguration"`
}

type ValidateOIDCSessionResponse struct {
	ValidateOIDCSession struct {
		Valid    bool   `json:"valid"`
		Username string `json:"username"`
	} `json:"validateOidcSession"`
}

type UpdateSettingsResponse struct {
	UpdateSettings struct {
		RestartRequired bool            `json:"restartRequired"`
		Warnings        []string        `json:"warnings"`
		Values          json.RawMessage `json:"values"`
	} `json:"updateSettings"`
}

type VM struct {
	Name  string `json:"name"`
	State string `json:"state"`
}
