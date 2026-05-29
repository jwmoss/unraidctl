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

var (
	arrayAddSlot    int
	arrayRemoveSlot int
)

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

		renderArray(resp.Array)
		return nil
	},
}

var arrayStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the array",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runArraySetState("START")
	},
}

var arrayStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the array",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runArraySetState("STOP")
	},
}

var arrayAddDiskCmd = &cobra.Command{
	Use:   "add-disk <disk-id>",
	Short: "Add a disk to the array",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := map[string]interface{}{"id": args[0]}
		if arrayAddSlot >= 0 {
			input["slot"] = arrayAddSlot
		}

		var resp api.ArrayMutationResponse
		if err := queryWithTimeout(api.ArrayAddDiskMutation, map[string]interface{}{"input": input}, &resp); err != nil {
			return fmt.Errorf("failed to add array disk: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.Array.AddDiskToArray)
		}
		out.Success("Disk added to array")
		renderArray(resp.Array.AddDiskToArray)
		return nil
	},
}

var arrayRemoveDiskCmd = &cobra.Command{
	Use:   "remove-disk <disk-id>",
	Short: "Remove a disk from the array",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := map[string]interface{}{"id": args[0]}
		if arrayRemoveSlot >= 0 {
			input["slot"] = arrayRemoveSlot
		}

		var resp api.ArrayMutationResponse
		if err := queryWithTimeout(api.ArrayRemoveDiskMutation, map[string]interface{}{"input": input}, &resp); err != nil {
			return fmt.Errorf("failed to remove array disk: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.Array.RemoveDiskFromArray)
		}
		out.Success("Disk removed from array")
		renderArray(resp.Array.RemoveDiskFromArray)
		return nil
	},
}

var arrayMountDiskCmd = &cobra.Command{
	Use:   "mount-disk <disk-id>",
	Short: "Mount an array disk",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runArrayDiskMutation(api.ArrayMountDiskMutation, args[0], "mountArrayDisk", "Disk mounted")
	},
}

var arrayUnmountDiskCmd = &cobra.Command{
	Use:   "unmount-disk <disk-id>",
	Short: "Unmount an array disk",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runArrayDiskMutation(api.ArrayUnmountDiskMutation, args[0], "unmountArrayDisk", "Disk unmounted")
	},
}

var arrayClearStatsCmd = &cobra.Command{
	Use:   "clear-stats <disk-id>",
	Short: "Clear array disk statistics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.ArrayMutationResponse
		if err := queryWithTimeout(api.ArrayClearDiskStatsMutation, map[string]interface{}{"id": args[0]}, &resp); err != nil {
			return fmt.Errorf("failed to clear disk statistics: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(map[string]bool{"cleared": resp.Array.ClearArrayDiskStatistics})
		}
		if resp.Array.ClearArrayDiskStatistics {
			out.Success("Disk statistics cleared")
		} else {
			out.Warn("Disk statistics were not cleared")
		}
		return nil
	},
}

func runArraySetState(state string) error {
	var resp api.ArrayMutationResponse
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"desiredState": state,
		},
	}
	if err := queryWithTimeout(api.ArraySetStateMutation, variables, &resp); err != nil {
		return fmt.Errorf("failed to set array state: %w", err)
	}

	if out.IsJSON() {
		return out.JSON(resp.Array.SetState)
	}
	out.Success(fmt.Sprintf("Array state set to %s", state))
	renderArray(resp.Array.SetState)
	return nil
}

func runArrayDiskMutation(query, id, fieldName, successMessage string) error {
	var resp api.ArrayMutationResponse
	if err := queryWithTimeout(query, map[string]interface{}{"id": id}, &resp); err != nil {
		return fmt.Errorf("failed to run %s: %w", fieldName, err)
	}

	var disk api.ArrayDisk
	switch fieldName {
	case "mountArrayDisk":
		disk = resp.Array.MountArrayDisk
	case "unmountArrayDisk":
		disk = resp.Array.UnmountArrayDisk
	}

	if out.IsJSON() {
		return out.JSON(disk)
	}
	out.Success(successMessage)
	renderArrayDisk(disk)
	return nil
}

func renderArray(arr api.Array) {
	out.Println("Array Status")
	out.Println("============")
	out.Print("State: %s\n", arr.State)

	// Capacity values are strings in KB in the Unraid API.
	if arr.Capacity.Kilobytes.Total != "" {
		total, _ := strconv.ParseInt(arr.Capacity.Kilobytes.Total, 10, 64)
		used, _ := strconv.ParseInt(arr.Capacity.Kilobytes.Used, 10, 64)
		free, _ := strconv.ParseInt(arr.Capacity.Kilobytes.Free, 10, 64)

		totalTB := float64(total) / (1024 * 1024 * 1024)
		usedTB := float64(used) / (1024 * 1024 * 1024)
		freeTB := float64(free) / (1024 * 1024 * 1024)
		usedPct := 0.0
		if total > 0 {
			usedPct = float64(used) / float64(total) * 100
		}

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
			rows = append(rows, formatArrayDiskRow(disk))
		}
		out.Table(headers, rows)
	}
}

func renderArrayDisk(disk api.ArrayDisk) {
	out.Table([]string{"NAME", "DEVICE", "TYPE", "SIZE", "STATUS", "TEMP"}, [][]string{formatArrayDiskRow(disk)})
}

func formatArrayDiskRow(disk api.ArrayDisk) []string {
	sizeTB := float64(disk.Size) / (1024 * 1024 * 1024)
	tempStr := "-"
	if disk.Temp > 0 {
		tempStr = fmt.Sprintf("%d°C", disk.Temp)
	}
	return []string{
		disk.Name,
		disk.Device,
		disk.Type,
		fmt.Sprintf("%.2f TB", sizeTB),
		disk.Status,
		tempStr,
	}
}

func init() {
	arrayCmd.AddCommand(arrayStatusCmd)
	arrayCmd.AddCommand(arrayStartCmd)
	arrayCmd.AddCommand(arrayStopCmd)
	arrayCmd.AddCommand(arrayAddDiskCmd)
	arrayCmd.AddCommand(arrayRemoveDiskCmd)
	arrayCmd.AddCommand(arrayMountDiskCmd)
	arrayCmd.AddCommand(arrayUnmountDiskCmd)
	arrayCmd.AddCommand(arrayClearStatsCmd)

	arrayAddDiskCmd.Flags().IntVar(&arrayAddSlot, "slot", -1, "array slot for the disk")
	arrayRemoveDiskCmd.Flags().IntVar(&arrayRemoveSlot, "slot", -1, "array slot for the disk")
}
