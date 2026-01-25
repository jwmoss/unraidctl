package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display system information",
	Long:  `Display system information including OS, CPU, and Unraid version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var resp api.InfoResponse
		if err := apiClient.Query(ctx, api.InfoQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get system info: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.Info)
		}

		info := resp.Info

		// Parse uptime from ISO 8601 boot timestamp
		var uptimeStr string
		if info.OS.Uptime != "" {
			bootTime, err := time.Parse(time.RFC3339, info.OS.Uptime)
			if err == nil {
				uptime := time.Since(bootTime)
				days := int(uptime.Hours()) / 24
				hours := int(uptime.Hours()) % 24
				mins := int(uptime.Minutes()) % 60

				if days > 0 {
					uptimeStr = fmt.Sprintf("%dd %dh %dm", days, hours, mins)
				} else if hours > 0 {
					uptimeStr = fmt.Sprintf("%dh %dm", hours, mins)
				} else {
					uptimeStr = fmt.Sprintf("%dm", mins)
				}
			} else {
				uptimeStr = info.OS.Uptime // fallback to raw value
			}
		}

		out.Println("System Information")
		out.Println("==================")
		out.Print("Hostname:  %s\n", info.OS.Hostname)
		out.Print("OS:        %s %s\n", info.OS.Distro, info.OS.Release)
		out.Print("Platform:  %s\n", info.OS.Platform)
		if uptimeStr != "" {
			out.Print("Uptime:    %s\n", uptimeStr)
		}

		if info.CPU.Brand != "" {
			out.Println("")
			out.Print("CPU:       %s %s\n", info.CPU.Manufacturer, info.CPU.Brand)
			out.Print("Cores:     %d\n", info.CPU.Cores)
			out.Print("Speed:     %.1f GHz\n", info.CPU.Speed)
		}

		return nil
	},
}
