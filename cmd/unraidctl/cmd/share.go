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
	Long:    `Commands for listing user shares on your Unraid server.`,
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

		headers := []string{"NAME", "USED", "FREE", "COMMENT"}
		var rows [][]string
		for _, s := range resp.Shares {
			usedStr := formatSizeKB(s.Used)
			freeStr := formatSizeKB(s.Free)

			rows = append(rows, []string{
				s.Name,
				usedStr,
				freeStr,
				s.Comment,
			})
		}
		out.Table(headers, rows)

		return nil
	},
}

func formatSizeKB(kb int64) string {
	if kb == 0 {
		return "-"
	}
	// Convert KB to appropriate unit
	const (
		KB = 1
		MB = 1024
		GB = 1024 * 1024
		TB = 1024 * 1024 * 1024
	)

	switch {
	case kb >= TB:
		return fmt.Sprintf("%.1f TB", float64(kb)/float64(TB))
	case kb >= GB:
		return fmt.Sprintf("%.1f GB", float64(kb)/float64(GB))
	case kb >= MB:
		return fmt.Sprintf("%.1f MB", float64(kb)/float64(MB))
	default:
		return fmt.Sprintf("%d KB", kb)
	}
}

func init() {
	shareCmd.AddCommand(shareListCmd)
}
