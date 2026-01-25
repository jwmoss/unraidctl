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
	Long:    `Commands for listing Docker containers on your Unraid server.`,
}

var dockerListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Docker containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var resp api.DockerResponse
		if err := apiClient.Query(ctx, api.DockerContainersQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list containers: %w", err)
		}

		containers := resp.Docker.Containers

		if out.IsJSON() {
			return out.JSON(containers)
		}

		if len(containers) == 0 {
			out.Println("No containers found")
			return nil
		}

		headers := []string{"NAME", "STATE", "STATUS", "IMAGE", "AUTOSTART"}
		var rows [][]string
		for _, c := range containers {
			name := "(unnamed)"
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
				truncate(c.Image, 40),
				autoStart,
			})
		}
		out.Table(headers, rows)

		return nil
	},
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	dockerCmd.AddCommand(dockerListCmd)
}
