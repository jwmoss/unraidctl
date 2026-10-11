package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var upsCmd = &cobra.Command{Use: "ups", Short: "View UPS status and configuration"}
var upsStatusCmd = &cobra.Command{
	Use: "status", Short: "Show raw UPS status and reported measurements", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.UPSResponse
		if err := queryWithTimeout(api.UPSQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get UPS status: %w", err)
		}
		return printJSONOr(resp.Devices, func() { renderUPS(resp.Devices) })
	},
}

var upsConfigCmd = &cobra.Command{
	Use: "config", Short: "Show the UPS daemon configuration", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp struct {
			Config map[string]interface{} `json:"upsConfiguration"`
		}
		if err := queryWithTimeout(api.UPSConfigurationQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get UPS configuration: %w", err)
		}
		return printJSONOr(resp.Config, func() { renderFields(resp.Config) })
	},
}

var upsSetCmd = &cobra.Command{
	// Not "configure": root.go treats that name as the local credentials command.
	Use:   "set",
	Short: "Change the UPS daemon configuration",
	Long:  "Change only the settings given as flags. Values use the API enums, for example --service ENABLE --cable USB --type USB --kill-ups NO.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := map[string]interface{}{}
		for flag, key := range map[string]string{"service": "service", "cable": "upsCable", "custom-cable": "customUpsCable", "type": "upsType", "device": "device", "kill-ups": "killUps"} {
			if cmd.Flags().Changed(flag) {
				value, _ := cmd.Flags().GetString(flag)
				if key != "device" && key != "customUpsCable" {
					value = strings.ToUpper(value) // API enums are uppercase.
				}
				config[key] = value
			}
		}
		for flag, key := range map[string]string{"capacity": "overrideUpsCapacity", "battery-level": "batteryLevel", "minutes": "minutes", "timeout-seconds": "timeout"} {
			if cmd.Flags().Changed(flag) {
				value, _ := cmd.Flags().GetInt(flag)
				config[key] = value
			}
		}
		if len(config) == 0 {
			return fmt.Errorf("set at least one setting flag")
		}
		if err := requireYes(cmd, "this changes when the server shuts down on battery power"); err != nil {
			return err
		}
		var resp struct {
			OK bool `json:"configureUps"`
		}
		if err := queryWithTimeout(api.ConfigureUPSMutation, map[string]interface{}{"config": config}, &resp); err != nil {
			return fmt.Errorf("failed to configure UPS: %w", err)
		}
		return reportResult(resp.OK, "UPS configuration updated", "UPS configuration was not updated")
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

var parityCmd = &cobra.Command{Use: "parity", Short: "View and control parity checks"}
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

var parityCorrect bool

var parityStartCmd = &cobra.Command{
	Use: "start", Short: "Start a parity check (read-only unless --correct)", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runParityMutation(api.ParityStartMutation, map[string]interface{}{"correct": parityCorrect}, "start")
	},
}
var parityPauseCmd = parityActionCommand("pause", "Pause the running parity check", api.ParityPauseMutation)
var parityResumeCmd = parityActionCommand("resume", "Resume a paused parity check", api.ParityResumeMutation)
var parityCancelCmd = parityActionCommand("cancel", "Cancel the running parity check", api.ParityCancelMutation)

func parityActionCommand(use, short, mutation string) *cobra.Command {
	return &cobra.Command{
		Use: use, Short: short, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return runParityMutation(mutation, nil, use) },
	}
}

// runParityMutation reports the untyped JSON result that the API returns for parity actions.
func runParityMutation(mutation string, variables map[string]interface{}, action string) error {
	var resp struct {
		ParityCheck map[string]json.RawMessage `json:"parityCheck"`
	}
	if err := queryWithTimeout(mutation, variables, &resp); err != nil {
		return fmt.Errorf("failed to %s parity check: %w", action, err)
	}
	result := resp.ParityCheck[action]
	return printJSONOr(result, func() { out.Success(fmt.Sprintf("Parity check %s requested: %s", action, result)) })
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
	upsCmd.AddCommand(upsStatusCmd, upsConfigCmd, upsSetCmd)
	diskCmd.AddCommand(diskListCmd)
	parityCmd.AddCommand(parityStatusCmd, parityHistoryCmd, parityStartCmd, parityPauseCmd, parityResumeCmd, parityCancelCmd)
	parityStartCmd.Flags().BoolVar(&parityCorrect, "correct", false, "write corrections to parity")
	for _, flag := range []string{"service", "cable", "custom-cable", "type", "device", "kill-ups"} {
		upsSetCmd.Flags().String(flag, "", "UPS "+flag+" setting")
	}
	for _, flag := range []string{"capacity", "battery-level", "minutes", "timeout-seconds"} {
		upsSetCmd.Flags().Int(flag, 0, "UPS "+flag+" setting")
	}
	upsSetCmd.Flags().Bool("yes", false, "confirm the change")
	rootCmd.AddCommand(upsCmd, diskCmd, parityCmd)
}
