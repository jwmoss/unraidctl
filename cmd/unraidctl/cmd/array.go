package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var arrayCmd = &cobra.Command{
	Use:   "array",
	Short: "Manage the disk array",
	Long:  `Commands for managing the Unraid disk array including status, start, and stop operations.`,
}

var arrayStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show array status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var resp api.ArrayResponse
		if err := apiClient.Query(ctx, api.ArrayStatusQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get array status: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.Array)
		}

		arr := resp.Array

		out.Println("Array Status")
		out.Println("============")
		out.Print("State: %s\n", arr.State)

		// Capacity (values are strings in KB)
		if arr.Capacity.Kilobytes.Total != "" {
			total, _ := strconv.ParseInt(arr.Capacity.Kilobytes.Total, 10, 64)
			used, _ := strconv.ParseInt(arr.Capacity.Kilobytes.Used, 10, 64)
			free, _ := strconv.ParseInt(arr.Capacity.Kilobytes.Free, 10, 64)

			totalTB := float64(total) / (1024 * 1024 * 1024) // KB to TB
			usedTB := float64(used) / (1024 * 1024 * 1024)
			freeTB := float64(free) / (1024 * 1024 * 1024)
			usedPct := float64(used) / float64(total) * 100

			out.Println("")
			out.Println("Capacity")
			out.Print("Total: %.2f TB\n", totalTB)
			out.Print("Used:  %.2f TB (%.1f%%)\n", usedTB, usedPct)
			out.Print("Free:  %.2f TB\n", freeTB)
		}

		if len(arr.Disks) > 0 {
			out.Println("")
			out.Println("Disks")
			headers := []string{"NAME", "DEVICE", "TYPE", "SIZE", "STATUS", "TEMP"}
			var rows [][]string
			for _, disk := range arr.Disks {
				sizeTB := float64(disk.Size) / (1024 * 1024 * 1024) // KB to TB
				tempStr := "-"
				if disk.Temp > 0 {
					tempStr = fmt.Sprintf("%d°C", disk.Temp)
				}
				rows = append(rows, []string{
					disk.Name,
					disk.Device,
					disk.Type,
					fmt.Sprintf("%.2f TB", sizeTB),
					disk.Status,
					tempStr,
				})
			}
			out.Table(headers, rows)
		}

		return nil
	},
}

func init() {
	arrayCmd.AddCommand(arrayStatusCmd)
}
