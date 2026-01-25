package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:     "share",
	Aliases: []string{"shares"},
	Short:   "Manage shares",
	Long:    `Commands for listing and managing user shares on your Unraid server.`,
}

var shareListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List shares",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var resp api.SharesResponse
		if err := apiClient.Query(ctx, api.SharesQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list shares: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.Shares)
		}

		if len(resp.Shares) == 0 {
			out.Println("No shares found")
			return nil
		}

		headers := []string{"NAME", "SIZE", "USED", "FREE", "COMMENT"}
		var rows [][]string
		for _, s := range resp.Shares {
			sizeStr := formatSize(s.Size)
			usedStr := formatSize(s.Used)
			freeStr := formatSize(s.Free)

			rows = append(rows, []string{
				s.Name,
				sizeStr,
				usedStr,
				freeStr,
				s.Comment,
			})
		}
		out.Table(headers, rows)

		return nil
	},
}

func formatSize(bytes int64) string {
	if bytes == 0 {
		return "-"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	shareCmd.AddCommand(shareListCmd)
}
