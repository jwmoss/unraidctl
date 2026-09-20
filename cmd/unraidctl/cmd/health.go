package cmd

import (
	"fmt"
	"strings"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

type healthCheck struct {
	Component string `json:"component"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type healthReport struct {
	Status string `json:"status"`
	api.InfoResponse
	api.ArrayResponse
	api.UPSResponse
	api.MetricsResponse
	api.NotificationsResponse
	api.ParityHistoryResponse
	api.DockerResponse
	Checks      []healthCheck `json:"checks"`
	Errors      []string      `json:"errors"`
	Limitations []string      `json:"limitations"`
}

var healthCmd = &cobra.Command{
	Use: "health", Short: "Check array, parity, UPS, containers, sensors, and unread alerts", Args: cobra.NoArgs,
	Long: "Show an API health snapshot. Exit 1 indicates an alert, warning, or incomplete data. Detailed SMART attributes and filesystem errors require separate diagnostics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		report := readHealth()
		if out.IsJSON() {
			if err := out.JSON(report); err != nil {
				return err
			}
		} else {
			renderHealth(report)
		}
		if report.Status != "OK" {
			return fmt.Errorf("health report contains alerts, warnings, or unavailable data")
		}
		return nil
	},
}

func readHealth() healthReport {
	report := healthReport{
		Errors: []string{},
		Limitations: []string{
			"API checks do not include detailed SMART attributes, Btrfs counters, scrub results, or kernel I/O errors.",
			"UPS battery condition is not independently verified; the API can supply default measurements.",
		},
	}
	// Separate requests preserve available checks when a resource is denied or unavailable.
	reads := []struct {
		name, query string
		result      interface{}
	}{
		{"system", api.InfoQuery, &report.InfoResponse},
		{"array", api.ArrayStatusQuery, &report.ArrayResponse},
		{"UPS", api.UPSQuery, &report.UPSResponse},
		{"metrics", api.SystemMetricsQuery, &report.MetricsResponse},
		{"notifications", api.NotificationAlertsQuery, &report.NotificationsResponse},
		{"parity history", api.ParityHistoryQuery, &report.ParityHistoryResponse},
		{"containers", api.ContainerHealthQuery, &report.DockerResponse},
	}
	for _, read := range reads {
		if err := queryWithTimeout(read.query, nil, read.result); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", read.name, err))
		}
	}
	report.Checks = assessHealth(report)
	report.Status = "OK"
	for _, check := range report.Checks {
		if check.Status != "OK" {
			report.Status = "ATTENTION"
		}
	}
	if len(report.Errors) > 0 {
		report.Status = "INCOMPLETE"
	}
	return report
}

func assessHealth(report healthReport) []healthCheck {
	checks := arrayHealthChecks(report.Array)
	for _, device := range report.Devices {
		checks = append(checks, healthCheck{"UPS: " + device.Name, upsStatusLevel(device.Status), device.Status})
	}
	if len(report.Devices) == 0 {
		checks = append(checks, healthCheck{"UPS", "UNKNOWN", "No UPS data available"})
	}
	checks = append(checks, parityHealthCheck(report.ParityHistory))
	counts := report.Notifications.Overview.Unread
	level := "OK"
	if counts.Warning > 0 {
		level = "WARNING"
	}
	if counts.Alert > 0 {
		level = "ALERT"
	}
	checks = append(checks, healthCheck{"Notifications", level, fmt.Sprintf("%d unread alerts; %d warnings; %d informational", counts.Alert, counts.Warning, counts.Info)})
	for _, container := range report.Docker.Containers {
		if strings.Contains(strings.ToLower(container.Status), "unhealthy") {
			checks = append(checks, healthCheck{"Container: " + containerName(container), "ALERT", container.Status})
		} else if container.AutoStart && container.State != "RUNNING" {
			checks = append(checks, healthCheck{"Container: " + containerName(container), "WARNING", container.State})
		}
	}
	if report.Metrics.Temperature != nil {
		for _, sensor := range report.Metrics.Temperature.Sensors {
			if sensor.Current.Status != "NORMAL" {
				checks = append(checks, healthCheck{"Temperature: " + sensor.Name, sensor.Current.Status, fmt.Sprintf("%.1f %s", sensor.Current.Value, sensor.Current.Unit)})
			}
		}
	} else {
		checks = append(checks, healthCheck{"Temperature", "UNKNOWN", "Temperature metrics are unavailable"})
	}
	return checks
}

func arrayHealthChecks(array api.Array) []healthCheck {
	level := "OK"
	if array.State != "STARTED" {
		level = "WARNING"
	}
	if array.State == "" {
		level = "UNKNOWN"
	}
	checks := []healthCheck{{"Array", level, array.State}}
	for _, disk := range array.AllDisks() {
		// Unassigned slots do not represent failed devices.
		if disk.Device == "" && disk.Size == 0 {
			continue
		}
		if disk.Status != "DISK_OK" {
			checks = append(checks, healthCheck{"Disk: " + disk.Name, "ALERT", disk.Status})
		}
		if disk.NumErrors != nil && *disk.NumErrors > 0 {
			checks = append(checks, healthCheck{"Disk: " + disk.Name, "ALERT", fmt.Sprintf("%d reported I/O errors", *disk.NumErrors)})
		}
	}
	return checks
}

func parityHealthCheck(history []api.ParityCheck) healthCheck {
	if len(history) == 0 {
		return healthCheck{"Parity", "UNKNOWN", "No parity check history available"}
	}
	latest := history[0]
	message := fmt.Sprintf("%s; %s; errors: %s", latest.Date, latest.Status, optionalInt(latest.Errors, ""))
	level := "OK"
	switch {
	case latest.Errors == nil:
		level = "UNKNOWN"
	case *latest.Errors > 0:
		level = "ALERT"
	case latest.Status != "COMPLETED":
		level = "WARNING"
	}
	return healthCheck{"Parity", level, message}
}

func upsStatusLevel(status string) string {
	flags := strings.Fields(strings.ToUpper(status))
	// Failure and communication flags take precedence over ONLINE.
	for _, flag := range flags {
		switch flag {
		case "REPLACEBATT", "NOBATT", "LOWBATT", "OVERLOAD", "SHUTTING":
			return "ALERT"
		case "COMMLOST":
			return "UNKNOWN"
		}
	}
	if strings.Contains(" "+strings.Join(flags, " ")+" ", " ONBATT ") {
		return "WARNING"
	}
	if len(flags) == 1 && flags[0] == "ONLINE" {
		return "OK"
	}
	return "UNKNOWN"
}

func renderHealth(report healthReport) {
	versions := report.Info.Versions.Core
	out.Print("Health: %s\nUnraid: %s | API: %s | Kernel: %s\n\n", report.Status, versions.Unraid, versions.API, versions.Kernel)
	rows := make([][]string, 0, len(report.Checks))
	for _, check := range report.Checks {
		rows = append(rows, []string{check.Component, check.Status, check.Message})
	}
	out.Table([]string{"CHECK", "STATUS", "DETAIL"}, rows)
	if report.Metrics.CPU != nil {
		out.Print("\nCPU: %.1f%%\n", report.Metrics.CPU.PercentTotal)
	}
	if report.Metrics.Memory != nil {
		out.Print("Available memory: %s\n", formatSizeBytes(report.Metrics.Memory.Available))
	}
	for _, err := range report.Errors {
		out.Print("Unavailable: %s\n", err)
	}
	for _, limit := range report.Limitations {
		out.Print("Scope: %s\n", limit)
	}
}

func init() { rootCmd.AddCommand(healthCmd) }
