package api

// Response types for Unraid API

type InfoResponse struct {
	Info struct {
		OS struct {
			Platform string `json:"platform"`
			Distro   string `json:"distro"`
			Release  string `json:"release"`
			Uptime   int64  `json:"uptime"`
			Hostname string `json:"hostname"`
		} `json:"os"`
		CPU struct {
			Manufacturer string  `json:"manufacturer"`
			Brand        string  `json:"brand"`
			Cores        int     `json:"cores"`
			Threads      int     `json:"threads"`
			Speed        float64 `json:"speed"`
		} `json:"cpu"`
		Memory struct {
			Total int64 `json:"total"`
			Free  int64 `json:"free"`
			Used  int64 `json:"used"`
		} `json:"memory"`
		Versions struct {
			Unraid string `json:"unraid"`
		} `json:"versions"`
	} `json:"info"`
}

type ArrayResponse struct {
	Array struct {
		State    string `json:"state"`
		Capacity struct {
			Disks struct {
				Free  int64 `json:"free"`
				Used  int64 `json:"used"`
				Total int64 `json:"total"`
			} `json:"disks"`
		} `json:"capacity"`
		Disks []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Size   int64  `json:"size"`
			Status string `json:"status"`
			Temp   int    `json:"temp"`
			Type   string `json:"type"`
		} `json:"disks"`
		ParityCheckProgress float64 `json:"parityCheckProgress"`
		ParityCheckRunning  bool    `json:"parityCheckRunning"`
	} `json:"array"`
}

type ArrayMutationResponse struct {
	ArrayStart *struct {
		State string `json:"state"`
	} `json:"arrayStart,omitempty"`
	ArrayStop *struct {
		State string `json:"state"`
	} `json:"arrayStop,omitempty"`
}

type DockerContainersResponse struct {
	DockerContainers []DockerContainer `json:"dockerContainers"`
}

type DockerContainer struct {
	ID        string   `json:"id"`
	Names     []string `json:"names"`
	State     string   `json:"state"`
	Status    string   `json:"status"`
	Image     string   `json:"image"`
	AutoStart bool     `json:"autoStart"`
}

type DockerMutationResponse struct {
	DockerContainerStart   *DockerContainer `json:"dockerContainerStart,omitempty"`
	DockerContainerStop    *DockerContainer `json:"dockerContainerStop,omitempty"`
	DockerContainerRestart *DockerContainer `json:"dockerContainerRestart,omitempty"`
}

type VMsResponse struct {
	VMs []VM `json:"vms"`
}

type VM struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	CoreCount int    `json:"coreCount"`
	Memory    int64  `json:"memory"`
}

type VMMutationResponse struct {
	VMStart *VM `json:"vmStart,omitempty"`
	VMStop  *VM `json:"vmStop,omitempty"`
}

type SharesResponse struct {
	Shares []Share `json:"shares"`
}

type Share struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Free    int64  `json:"free"`
	Used    int64  `json:"used"`
	Size    int64  `json:"size"`
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
}

type Notification struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Importance  string `json:"importance"`
	Timestamp   string `json:"timestamp"`
}
