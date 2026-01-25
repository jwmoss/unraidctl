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

var forceVM bool

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Manage virtual machines",
	Long:  `Commands for listing and managing virtual machines on your Unraid server.`,
}

var vmListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List virtual machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var resp api.VMsResponse
		if err := apiClient.Query(ctx, api.VMsQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list VMs: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.VMs)
		}

		if len(resp.VMs) == 0 {
			out.Println("No VMs found")
			return nil
		}

		headers := []string{"NAME", "STATE", "CORES", "MEMORY"}
		var rows [][]string
		for _, vm := range resp.VMs {
			memGB := float64(vm.Memory) / (1024 * 1024 * 1024)
			rows = append(rows, []string{
				vm.Name,
				vm.State,
				fmt.Sprintf("%d", vm.CoreCount),
				fmt.Sprintf("%.1f GB", memGB),
			})
		}
		out.Table(headers, rows)

		return nil
	},
}

var vmStartCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		vars := map[string]interface{}{"id": vmName}
		var resp api.VMMutationResponse
		if err := apiClient.Query(ctx, api.VMStartMutation, vars, &resp); err != nil {
			return fmt.Errorf("failed to start VM: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp)
		}

		out.Success(fmt.Sprintf("VM %s started", vmName))
		return nil
	},
}

var vmStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a virtual machine",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]

		if !forceVM {
			out.Print("Are you sure you want to stop VM %s? [y/N]: ", vmName)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				out.Println("Aborted.")
				return nil
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		vars := map[string]interface{}{"id": vmName}
		var resp api.VMMutationResponse
		if err := apiClient.Query(ctx, api.VMStopMutation, vars, &resp); err != nil {
			return fmt.Errorf("failed to stop VM: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp)
		}

		out.Success(fmt.Sprintf("VM %s stopped", vmName))
		return nil
	},
}

func init() {
	vmCmd.AddCommand(vmListCmd)
	vmCmd.AddCommand(vmStartCmd)
	vmCmd.AddCommand(vmStopCmd)

	vmStopCmd.Flags().BoolVarP(&forceVM, "force", "f", false, "skip confirmation prompt")
}
