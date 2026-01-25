package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:     "docker",
	Aliases: []string{"container"},
	Short:   "Manage Docker containers",
	Long:    `Commands for listing and managing Docker containers on your Unraid server.`,
}

var dockerListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Docker containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var resp api.DockerContainersResponse
		if err := apiClient.Query(ctx, api.DockerContainersQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list containers: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.DockerContainers)
		}

		if len(resp.DockerContainers) == 0 {
			out.Println("No containers found")
			return nil
		}

		headers := []string{"NAME", "STATE", "STATUS", "IMAGE", "AUTOSTART"}
		var rows [][]string
		for _, c := range resp.DockerContainers {
			name := c.ID[:12]
			if len(c.Names) > 0 {
				name = strings.TrimPrefix(c.Names[0], "/")
			}
			autoStart := "no"
			if c.AutoStart {
				autoStart = "yes"
			}
			rows = append(rows, []string{
				name,
				c.State,
				c.Status,
				c.Image,
				autoStart,
			})
		}
		out.Table(headers, rows)

		return nil
	},
}

var dockerStartCmd = &cobra.Command{
	Use:   "start <container>",
	Short: "Start a Docker container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerID := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		vars := map[string]interface{}{"id": containerID}
		var resp api.DockerMutationResponse
		if err := apiClient.Query(ctx, api.DockerStartMutation, vars, &resp); err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp)
		}

		out.Success(fmt.Sprintf("Container %s started", containerID))
		return nil
	},
}

var dockerStopCmd = &cobra.Command{
	Use:   "stop <container>",
	Short: "Stop a Docker container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerID := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		vars := map[string]interface{}{"id": containerID}
		var resp api.DockerMutationResponse
		if err := apiClient.Query(ctx, api.DockerStopMutation, vars, &resp); err != nil {
			return fmt.Errorf("failed to stop container: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp)
		}

		out.Success(fmt.Sprintf("Container %s stopped", containerID))
		return nil
	},
}

var dockerRestartCmd = &cobra.Command{
	Use:   "restart <container>",
	Short: "Restart a Docker container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerID := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		vars := map[string]interface{}{"id": containerID}
		var resp api.DockerMutationResponse
		if err := apiClient.Query(ctx, api.DockerRestartMutation, vars, &resp); err != nil {
			return fmt.Errorf("failed to restart container: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp)
		}

		out.Success(fmt.Sprintf("Container %s restarted", containerID))
		return nil
	},
}

func init() {
	dockerCmd.AddCommand(dockerListCmd)
	dockerCmd.AddCommand(dockerStartCmd)
	dockerCmd.AddCommand(dockerStopCmd)
	dockerCmd.AddCommand(dockerRestartCmd)
}
