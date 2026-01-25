package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var forceArray bool

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

		if arr.ParityCheckRunning {
			out.Print("Parity Check: %.1f%% complete\n", arr.ParityCheckProgress)
		}

		// Capacity
		totalTB := float64(arr.Capacity.Disks.Total) / (1024 * 1024 * 1024 * 1024)
		usedTB := float64(arr.Capacity.Disks.Used) / (1024 * 1024 * 1024 * 1024)
		freeTB := float64(arr.Capacity.Disks.Free) / (1024 * 1024 * 1024 * 1024)
		usedPct := float64(arr.Capacity.Disks.Used) / float64(arr.Capacity.Disks.Total) * 100

		out.Println("")
		out.Println("Capacity")
		out.Print("Total: %.2f TB\n", totalTB)
		out.Print("Used:  %.2f TB (%.1f%%)\n", usedTB, usedPct)
		out.Print("Free:  %.2f TB\n", freeTB)

		if len(arr.Disks) > 0 {
			out.Println("")
			out.Println("Disks")
			headers := []string{"NAME", "TYPE", "SIZE", "STATUS", "TEMP"}
			var rows [][]string
			for _, disk := range arr.Disks {
				sizeTB := float64(disk.Size) / (1024 * 1024 * 1024 * 1024)
				tempStr := "-"
				if disk.Temp > 0 {
					tempStr = fmt.Sprintf("%d°C", disk.Temp)
				}
				rows = append(rows, []string{
					disk.Name,
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

var arrayStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the array",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !forceArray {
			out.Print("Are you sure you want to start the array? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				out.Println("Aborted.")
				return nil
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var resp api.ArrayMutationResponse
		if err := apiClient.Query(ctx, api.ArrayStartMutation, nil, &resp); err != nil {
			return fmt.Errorf("failed to start array: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp)
		}

		out.Success("Array started")
		return nil
	},
}

var arrayStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the array",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !forceArray {
			out.Warn("Stopping the array will make all data inaccessible!")
			out.Print("Are you sure you want to stop the array? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				out.Println("Aborted.")
				return nil
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var resp api.ArrayMutationResponse
		if err := apiClient.Query(ctx, api.ArrayStopMutation, nil, &resp); err != nil {
			return fmt.Errorf("failed to stop array: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp)
		}

		out.Success("Array stopped")
		return nil
	},
}

func init() {
	arrayCmd.AddCommand(arrayStatusCmd)
	arrayCmd.AddCommand(arrayStartCmd)
	arrayCmd.AddCommand(arrayStopCmd)

	arrayStartCmd.Flags().BoolVarP(&forceArray, "force", "f", false, "skip confirmation prompt")
	arrayStopCmd.Flags().BoolVarP(&forceArray, "force", "f", false, "skip confirmation prompt")
}
