package api

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
	Array struct {
		State    string `json:"state"`
		Capacity struct {
			Kilobytes struct {
				Free  string `json:"free"` // String in API response
				Used  string `json:"used"`
				Total string `json:"total"`
			} `json:"kilobytes"`
		} `json:"capacity"`
		Disks []ArrayDisk `json:"disks"`
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
	ID        string   `json:"id"`
	Names     []string `json:"names"`
	State     string   `json:"state"`
	Status    string   `json:"status"`
	Image     string   `json:"image"`
	AutoStart bool     `json:"autoStart"`
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

type VM struct {
	Name  string `json:"name"`
	State string `json:"state"`
}
