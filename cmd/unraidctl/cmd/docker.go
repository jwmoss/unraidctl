package cmd

import (
	"fmt"
	"strings"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:     "docker",
	Aliases: []string{"container"},
	Short:   "Manage Docker containers",
	Long:    `Commands for listing Docker containers on your Unraid server.`,
}

var (
	dockerWide             bool
	dockerLogsTail         int
	dockerLogsSince        string
	dockerRemoveWithImage  bool
	dockerAutostartEnable  bool
	dockerAutostartDisable bool
	dockerAutostartWait    int
	dockerAutostartPersist bool
)

var dockerListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Docker containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.DockerResponse
		if err := queryWithTimeout(api.DockerContainersQuery, nil, &resp); err != nil {
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

		renderDockerContainers(containers, dockerWide)
		return nil
	},
}

var dockerInspectCmd = &cobra.Command{
	Use:   "inspect <container-id>",
	Short: "Show detailed Docker container metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.DockerContainerResponse
		if err := queryWithTimeout(api.DockerContainerQuery, map[string]interface{}{"id": args[0]}, &resp); err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}

		container := resp.Docker.Container
		if container.ID == "" {
			return fmt.Errorf("container not found: %s", args[0])
		}

		if out.IsJSON() {
			return out.JSON(container)
		}
		renderDockerContainerDetails(container)
		return nil
	},
}

var dockerLogsCmd = &cobra.Command{
	Use:   "logs <container-id>",
	Short: "Show Docker container logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		variables := map[string]interface{}{"id": args[0]}
		if dockerLogsTail > 0 {
			variables["tail"] = dockerLogsTail
		}
		if dockerLogsSince != "" {
			variables["since"] = dockerLogsSince
		}

		var resp api.DockerLogsResponse
		if err := queryWithTimeout(api.DockerLogsQuery, variables, &resp); err != nil {
			return fmt.Errorf("failed to read container logs: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.Docker.Logs)
		}
		for _, line := range resp.Docker.Logs.Lines {
			if line.Timestamp == "" {
				out.Println(line.Message)
				continue
			}
			out.Print("%s %s\n", line.Timestamp, line.Message)
		}
		return nil
	},
}

var dockerStartCmd = dockerActionCommand("start", "Start a Docker container", api.DockerStartMutation)
var dockerStopCmd = dockerActionCommand("stop", "Stop a Docker container", api.DockerStopMutation)
var dockerRestartCmd = dockerActionCommand("restart", "Restart a Docker container", api.DockerRestartMutation)
var dockerPauseCmd = dockerActionCommand("pause", "Pause a Docker container", api.DockerPauseMutation)
var dockerUnpauseCmd = dockerActionCommand("unpause", "Unpause a Docker container", api.DockerUnpauseMutation)
var dockerUpdateCmd = dockerActionCommand("update", "Update a Docker container image", api.DockerUpdateMutation)

var dockerUpdateAllCmd = &cobra.Command{
	Use:   "update-all",
	Short: "Update every Docker container with an available update",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.DockerMutationResponse
		if err := queryWithTimeout(api.DockerUpdateAllMutation, nil, &resp); err != nil {
			return fmt.Errorf("failed to update containers: %w", err)
		}
		containers := resp.Docker.UpdateAllContainers
		if out.IsJSON() {
			return out.JSON(containers)
		}
		out.Success(fmt.Sprintf("Updated %d container(s)", len(containers)))
		if len(containers) > 0 {
			renderDockerContainers(containers, true)
		}
		return nil
	},
}

var dockerRemoveCmd = &cobra.Command{
	Use:   "remove <container-id>",
	Short: "Remove a Docker container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.DockerMutationResponse
		variables := map[string]interface{}{
			"id":        args[0],
			"withImage": dockerRemoveWithImage,
		}
		if err := queryWithTimeout(api.DockerRemoveMutation, variables, &resp); err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(map[string]bool{"removed": resp.Docker.RemoveContainer})
		}
		if resp.Docker.RemoveContainer {
			out.Success("Container removed")
		} else {
			out.Warn("Container was not removed")
		}
		return nil
	},
}

var dockerAutostartCmd = &cobra.Command{
	Use:   "autostart <container-id>",
	Short: "Update Docker autostart settings",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dockerAutostartEnable == dockerAutostartDisable {
			return fmt.Errorf("set exactly one of --enable or --disable")
		}

		entry := map[string]interface{}{
			"id":        args[0],
			"autoStart": dockerAutostartEnable,
		}
		if dockerAutostartWait >= 0 {
			entry["wait"] = dockerAutostartWait
		}
		variables := map[string]interface{}{
			"entries": []map[string]interface{}{entry},
			"persist": dockerAutostartPersist,
		}

		var resp api.DockerMutationResponse
		if err := queryWithTimeout(api.DockerAutostartMutation, variables, &resp); err != nil {
			return fmt.Errorf("failed to update Docker autostart: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(map[string]bool{"updated": resp.Docker.UpdateAutostartConfiguration})
		}
		if resp.Docker.UpdateAutostartConfiguration {
			out.Success("Docker autostart updated")
		} else {
			out.Warn("Docker autostart was not updated")
		}
		return nil
	},
}

func dockerActionCommand(use, short, query string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <container-id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp api.DockerMutationResponse
			if err := queryWithTimeout(query, map[string]interface{}{"id": args[0]}, &resp); err != nil {
				return fmt.Errorf("failed to %s container: %w", use, err)
			}

			container := dockerActionContainer(use, resp)
			if out.IsJSON() {
				return out.JSON(container)
			}
			out.Success(fmt.Sprintf("Container %s complete", use))
			renderDockerContainers([]api.DockerContainer{container}, true)
			return nil
		},
	}
}

func dockerActionContainer(action string, resp api.DockerMutationResponse) api.DockerContainer {
	switch action {
	case "start":
		return resp.Docker.Start
	case "stop":
		return resp.Docker.Stop
	case "restart":
		return resp.Docker.Restart
	case "pause":
		return resp.Docker.Pause
	case "unpause":
		return resp.Docker.Unpause
	case "update":
		return resp.Docker.UpdateContainer
	default:
		return api.DockerContainer{}
	}
}

func renderDockerContainers(containers []api.DockerContainer, wide bool) {
	headers := []string{"NAME", "STATE", "STATUS", "IMAGE", "AUTOSTART"}
	if wide {
		headers = []string{"NAME", "STATE", "STATUS", "IMAGE", "AUTOSTART", "UPDATE", "WEB UI", "PORTS"}
	}

	var rows [][]string
	for _, c := range containers {
		autoStart := "no"
		if c.AutoStart {
			autoStart = "yes"
		}

		row := []string{
			containerName(c),
			c.State,
			c.Status,
			truncate(c.Image, 40),
			autoStart,
		}
		if wide {
			row = append(row,
				updateStatus(c),
				c.WebUIURL,
				truncate(strings.Join(c.LanIPPorts, ","), 40),
			)
		}
		rows = append(rows, row)
	}
	out.Table(headers, rows)
}

func renderDockerContainerDetails(c api.DockerContainer) {
	rows := [][]string{
		{"ID", c.ID},
		{"Name", containerName(c)},
		{"State", c.State},
		{"Status", c.Status},
		{"Image", c.Image},
		{"Image ID", c.ImageID},
		{"Command", c.Command},
		{"Autostart", fmt.Sprintf("%t", c.AutoStart)},
		{"Update", updateStatus(c)},
		{"Orphaned", fmt.Sprintf("%t", c.IsOrphaned)},
		{"Tailscale", fmt.Sprintf("%t", c.TailscaleEnabled)},
		{"Web UI", c.WebUIURL},
		{"Project", c.ProjectURL},
		{"Registry", c.RegistryURL},
		{"Support", c.SupportURL},
		{"Icon", c.IconURL},
		{"Shell", c.Shell},
		{"LAN ports", strings.Join(c.LanIPPorts, ", ")},
	}
	if c.HostConfig != nil {
		rows = append(rows, []string{"Network mode", c.HostConfig.NetworkMode})
	}
	if c.AutoStartOrder != nil {
		rows = append(rows, []string{"Autostart order", fmt.Sprintf("%d", *c.AutoStartOrder)})
	}
	if c.AutoStartWait != nil {
		rows = append(rows, []string{"Autostart wait", fmt.Sprintf("%ds", *c.AutoStartWait)})
	}
	if c.SizeRootFs != nil {
		rows = append(rows, []string{"Root FS size", formatSizeBytes(*c.SizeRootFs)})
	}
	if c.SizeRw != nil {
		rows = append(rows, []string{"Writable size", formatSizeBytes(*c.SizeRw)})
	}
	if c.SizeLog != nil {
		rows = append(rows, []string{"Log size", formatSizeBytes(*c.SizeLog)})
	}
	out.Table([]string{"FIELD", "VALUE"}, rows)
}

func containerName(c api.DockerContainer) string {
	if len(c.Names) == 0 {
		return "(unnamed)"
	}
	return strings.TrimPrefix(c.Names[0], "/")
}

func updateStatus(c api.DockerContainer) string {
	switch {
	case c.IsRebuildReady != nil && *c.IsRebuildReady:
		return "rebuild ready"
	case c.IsUpdateAvailable != nil && *c.IsUpdateAvailable:
		return "available"
	case c.IsUpdateAvailable != nil:
		return "current"
	default:
		return "-"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	dockerCmd.AddCommand(dockerListCmd)
	dockerCmd.AddCommand(dockerInspectCmd)
	dockerCmd.AddCommand(dockerLogsCmd)
	dockerCmd.AddCommand(dockerStartCmd)
	dockerCmd.AddCommand(dockerStopCmd)
	dockerCmd.AddCommand(dockerRestartCmd)
	dockerCmd.AddCommand(dockerPauseCmd)
	dockerCmd.AddCommand(dockerUnpauseCmd)
	dockerCmd.AddCommand(dockerUpdateCmd)
	dockerCmd.AddCommand(dockerUpdateAllCmd)
	dockerCmd.AddCommand(dockerRemoveCmd)
	dockerCmd.AddCommand(dockerAutostartCmd)

	dockerListCmd.Flags().BoolVarP(&dockerWide, "wide", "w", false, "show extended Docker metadata")
	dockerLogsCmd.Flags().IntVar(&dockerLogsTail, "tail", 100, "number of log lines to read")
	dockerLogsCmd.Flags().StringVar(&dockerLogsSince, "since", "", "read logs since this RFC3339 timestamp")
	dockerRemoveCmd.Flags().BoolVar(&dockerRemoveWithImage, "with-image", false, "remove the backing image too")
	dockerAutostartCmd.Flags().BoolVar(&dockerAutostartEnable, "enable", false, "enable autostart")
	dockerAutostartCmd.Flags().BoolVar(&dockerAutostartDisable, "disable", false, "disable autostart")
	dockerAutostartCmd.Flags().IntVar(&dockerAutostartWait, "wait", -1, "seconds to wait after autostarting")
	dockerAutostartCmd.Flags().BoolVar(&dockerAutostartPersist, "persist", true, "persist user preferences")
}
