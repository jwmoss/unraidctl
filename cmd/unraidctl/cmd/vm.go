package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Manage virtual machines",
	Long:  `Commands for listing virtual machines on your Unraid server.`,
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
			if strings.Contains(err.Error(), "Failed to retrieve VM domains: VMs are not available") {
				return fmt.Errorf("VM Manager is disabled or unavailable on this server: %w", err)
			}
			return fmt.Errorf("failed to list VMs: %w", err)
		}

		vms := resp.VMs.Domain

		if out.IsJSON() {
			return out.JSON(vms)
		}

		if len(vms) == 0 {
			out.Println("No VMs found (VMs may not be enabled on this server)")
			return nil
		}

		headers := []string{"NAME", "STATE"}
		var rows [][]string
		for _, vm := range vms {
			rows = append(rows, []string{
				vm.Name,
				vm.State,
			})
		}
		out.Table(headers, rows)

		return nil
	},
}

func init() {
	vmCmd.AddCommand(vmListCmd)
}
