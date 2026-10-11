package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Manage virtual machines",
	Long:  `Commands for listing and controlling virtual machines on your Unraid server.`,
}

var vmListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List virtual machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		var resp api.VMsResponse
		if err := apiClient.Query(ctx, api.VMListQuery, nil, &resp); err != nil {
			if strings.Contains(err.Error(), "Failed to retrieve VM domains: VMs are not available") {
				return fmt.Errorf("VM Manager is disabled or unavailable on this server: %w", err)
			}
			return fmt.Errorf("failed to list VMs: %w", err)
		}

		vms := resp.VMs.Domains

		if out.IsJSON() {
			return out.JSON(vms)
		}

		if len(vms) == 0 {
			out.Println("No VMs found (VMs may not be enabled on this server)")
			return nil
		}

		headers := []string{"ID", "NAME", "STATE"}
		var rows [][]string
		for _, vm := range vms {
			rows = append(rows, []string{vm.ID, vm.Name, vm.State})
		}
		out.Table(headers, rows)

		return nil
	},
}

func vmActionCommand(use, field, short, mutation string) *cobra.Command {
	return &cobra.Command{
		Use: use + " <vm-id>", Short: short, Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp struct {
				VM map[string]bool `json:"vm"`
			}
			if err := queryWithTimeout(mutation, map[string]interface{}{"id": args[0]}, &resp); err != nil {
				return fmt.Errorf("failed to %s VM: %w", use, err)
			}
			return reportResult(resp.VM[field], "VM "+use+" complete", "VM "+use+" was not applied")
		},
	}
}

func init() {
	vmCmd.AddCommand(vmListCmd,
		vmActionCommand("start", "start", "Start a VM", api.VMStartMutation),
		vmActionCommand("stop", "stop", "Shut down a VM gracefully", api.VMStopMutation),
		vmActionCommand("pause", "pause", "Pause a VM", api.VMPauseMutation),
		vmActionCommand("resume", "resume", "Resume a paused VM", api.VMResumeMutation),
		vmActionCommand("force-stop", "forceStop", "Power off a VM immediately", api.VMForceStopMutation),
		vmActionCommand("reboot", "reboot", "Reboot a VM gracefully", api.VMRebootMutation),
		vmActionCommand("reset", "reset", "Hard-reset a VM", api.VMResetMutation),
	)
}
