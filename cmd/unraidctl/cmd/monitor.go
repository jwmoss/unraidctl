package cmd

import (
	"fmt"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var upsCmd = &cobra.Command{Use: "ups", Short: "View UPS status"}
var upsStatusCmd = &cobra.Command{
	Use: "status", Short: "Show raw UPS status and reported measurements", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.UPSResponse
		if err := queryWithTimeout(api.UPSQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get UPS status: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.Devices)
		}
		renderUPS(resp.Devices)
		return nil
	},
}

func renderUPS(devices []api.UPSDevice) {
	if len(devices) == 0 {
		out.Println("No UPS data available")
		return
	}
	rows := make([][]string, 0, len(devices))
	for _, device := range devices {
		rows = append(rows, []string{device.Name, device.Status, optionalInt(device.Battery.ChargeLevel, "%"), optionalInt(device.Battery.EstimatedRuntime, " s"), optionalInt(device.Power.LoadPercentage, "%")})
	}
	out.Table([]string{"NAME", "STATUS", "CHARGE", "ESTIMATED RUNTIME", "LOAD"}, rows)
	out.Println("Use raw status for alarms. API battery measurements can include defaults; battery condition is not independently verified.")
}

func optionalInt(value *int, suffix string) string {
	if value == nil {
		return "unknown"
	}
	return fmt.Sprintf("%d%s", *value, suffix)
}

var diskCmd = &cobra.Command{Use: "disk", Short: "View physical disks"}
var diskListCmd = &cobra.Command{
	Use: "list", Aliases: []string{"ls"}, Short: "List physical disks and basic SMART status", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.DisksResponse
		if err := queryWithTimeout(api.DisksQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get disks: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.Disks)
		}
		rows := make([][]string, 0, len(resp.Disks))
		for _, disk := range resp.Disks {
			temp := "unknown"
			if disk.Temperature != nil {
				temp = fmt.Sprintf("%.0f C", *disk.Temperature)
			}
			rows = append(rows, []string{disk.Device, disk.Name, disk.Type, formatSizeBytes(disk.Size), disk.SmartStatus, temp})
		}
		out.Table([]string{"DEVICE", "MODEL", "TYPE", "CAPACITY", "SMART", "TEMPERATURE"}, rows)
		out.Println("Basic SMART status does not include detailed attributes or filesystem error counters.")
		return nil
	},
}

var parityCmd = &cobra.Command{Use: "parity", Short: "View parity checks"}
var parityStatusCmd = &cobra.Command{
	Use: "status", Short: "Show the current or latest parity check", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.ArrayResponse
		if err := queryWithTimeout(api.ParityStatusQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get parity status: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.Array.ParityCheckStatus)
		}
		if resp.Array.ParityCheckStatus == nil {
			return fmt.Errorf("parity status is unavailable")
		}
		renderParity([]api.ParityCheck{*resp.Array.ParityCheckStatus})
		out.Println("Use parity history for completed-check error counts when status reports unknown.")
		return nil
	},
}
var parityHistoryCmd = &cobra.Command{
	Use: "history", Short: "List parity check history", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.ParityHistoryResponse
		if err := queryWithTimeout(api.ParityHistoryQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get parity history: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.ParityHistory)
		}
		renderParity(resp.ParityHistory)
		return nil
	},
}

func renderParity(checks []api.ParityCheck) {
	if len(checks) == 0 {
		out.Println("No parity check history")
		return
	}
	rows := make([][]string, 0, len(checks))
	for _, check := range checks {
		rows = append(rows, []string{check.Date, check.Status, optionalInt(check.Errors, ""), optionalInt(check.Progress, "%"), optionalInt(check.Duration, " s")})
	}
	out.Table([]string{"DATE", "STATUS", "ERRORS", "PROGRESS", "DURATION"}, rows)
}

func init() {
	upsCmd.AddCommand(upsStatusCmd)
	diskCmd.AddCommand(diskListCmd)
	parityCmd.AddCommand(parityStatusCmd, parityHistoryCmd)
	rootCmd.AddCommand(upsCmd, diskCmd, parityCmd)
}
