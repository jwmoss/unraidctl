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
	Long:  `Display detailed system information including OS, CPU, memory, and Unraid version.`,
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

		// Format uptime
		uptime := time.Duration(info.OS.Uptime) * time.Second
		days := int(uptime.Hours()) / 24
		hours := int(uptime.Hours()) % 24
		mins := int(uptime.Minutes()) % 60

		var uptimeStr string
		if days > 0 {
			uptimeStr = fmt.Sprintf("%dd %dh %dm", days, hours, mins)
		} else if hours > 0 {
			uptimeStr = fmt.Sprintf("%dh %dm", hours, mins)
		} else {
			uptimeStr = fmt.Sprintf("%dm", mins)
		}

		// Format memory
		totalGB := float64(info.Memory.Total) / (1024 * 1024 * 1024)
		usedGB := float64(info.Memory.Used) / (1024 * 1024 * 1024)
		usedPct := float64(info.Memory.Used) / float64(info.Memory.Total) * 100

		out.Println("System Information")
		out.Println("==================")
		out.Print("Hostname:      %s\n", info.OS.Hostname)
		out.Print("Unraid:        %s\n", info.Versions.Unraid)
		out.Print("OS:            %s %s\n", info.OS.Distro, info.OS.Release)
		out.Print("Uptime:        %s\n", uptimeStr)
		out.Println("")
		out.Print("CPU:           %s %s\n", info.CPU.Manufacturer, info.CPU.Brand)
		out.Print("Cores/Threads: %d / %d\n", info.CPU.Cores, info.CPU.Threads)
		out.Println("")
		out.Print("Memory:        %.1f / %.1f GB (%.1f%%)\n", usedGB, totalGB, usedPct)

		return nil
	},
}
