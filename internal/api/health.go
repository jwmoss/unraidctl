package api

// CoreVersions reports exact component versions, unlike info.os.release.
type CoreVersions struct {
	Unraid string `json:"unraid"`
	API    string `json:"api"`
	Kernel string `json:"kernel"`
}

func (a Array) AllDisks() []ArrayDisk {
	disks := make([]ArrayDisk, 0, len(a.Disks)+len(a.Parities)+len(a.Caches)+1)
	disks = append(disks, a.Parities...)
	disks = append(disks, a.Disks...)
	disks = append(disks, a.Caches...)
	if a.Boot != nil {
		disks = append(disks, *a.Boot)
	}
	return disks
}

type ParityCheck struct {
	Date       string `json:"date"`
	Duration   *int   `json:"duration"`
	Status     string `json:"status"`
	Errors     *int   `json:"errors"`
	Progress   *int   `json:"progress"`
	Running    *bool  `json:"running"`
	Paused     *bool  `json:"paused"`
	Correcting *bool  `json:"correcting"`
}

type ParityHistoryResponse struct {
	ParityHistory []ParityCheck `json:"parityHistory"`
}

type DisksResponse struct {
	Disks []Disk `json:"disks"`
}

type Disk struct {
	ID          string   `json:"id"`
	Device      string   `json:"device"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Size        int64    `json:"size"` // Bytes.
	SmartStatus string   `json:"smartStatus"`
	Temperature *float64 `json:"temperature"`
}

type UPSResponse struct {
	Devices []UPSDevice `json:"upsDevices"`
}

type UPSDevice struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	Status  string `json:"status"`
	Battery struct {
		ChargeLevel      *int `json:"chargeLevel"`
		EstimatedRuntime *int `json:"estimatedRuntime"` // Seconds.
		// Do not request battery.health: API 4.37.x hard-codes it to Good.
	} `json:"battery"`
	Power struct {
		LoadPercentage *int     `json:"loadPercentage"`
		CurrentPower   *float64 `json:"currentPower"`
	} `json:"power"`
}

type Metrics struct {
	Network     []NetworkMetrics    `json:"network,omitempty"`
	CPU         *CPUUtilization     `json:"cpu,omitempty"`
	Memory      *MemoryUtilization  `json:"memory,omitempty"`
	Temperature *TemperatureMetrics `json:"temperature,omitempty"`
}

type CPUUtilization struct {
	PercentTotal float64 `json:"percentTotal"`
}

type MemoryUtilization struct {
	Total        int64   `json:"total"`
	Available    int64   `json:"available"`
	Active       int64   `json:"active"`
	PercentTotal float64 `json:"percentTotal"`
}

type TemperatureMetrics struct {
	Sensors []TemperatureSensor `json:"sensors"`
}

type TemperatureSensor struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Current struct {
		Value  float64 `json:"value"`
		Unit   string  `json:"unit"`
		Status string  `json:"status"`
	} `json:"current"`
}
