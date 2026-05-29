package cmd

import (
	"fmt"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Aliases: []string{"logs"},
	Short:   "Read Unraid API log files",
}

var (
	logShowLines     int
	logShowStartLine int
)

var logListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available log files",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.LogFilesResponse
		if err := queryWithTimeout(api.LogFilesQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list log files: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.LogFiles)
		}
		rows := make([][]string, 0, len(resp.LogFiles))
		for _, logFile := range resp.LogFiles {
			rows = append(rows, []string{
				logFile.Name,
				logFile.Path,
				formatSizeBytes(logFile.Size),
				logFile.ModifiedAt,
			})
		}
		out.Table([]string{"NAME", "PATH", "SIZE", "MODIFIED"}, rows)
		return nil
	},
}

var logShowCmd = &cobra.Command{
	Use:     "show <path>",
	Aliases: []string{"cat", "tail"},
	Short:   "Show a log file",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		variables := map[string]interface{}{"path": args[0]}
		if logShowLines > 0 {
			variables["lines"] = logShowLines
		}
		if logShowStartLine > 0 {
			variables["startLine"] = logShowStartLine
		}

		var resp api.LogFileResponse
		if err := queryWithTimeout(api.LogFileQuery, variables, &resp); err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.LogFile)
		}
		out.Print("%s", resp.LogFile.Content)
		return nil
	},
}

func init() {
	logCmd.AddCommand(logListCmd)
	logCmd.AddCommand(logShowCmd)

	logShowCmd.Flags().IntVar(&logShowLines, "lines", 100, "number of lines to read")
	logShowCmd.Flags().IntVar(&logShowStartLine, "start-line", 0, "1-indexed start line")
}
