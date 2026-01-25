package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var showAll bool

var notificationCmd = &cobra.Command{
	Use:     "notification",
	Aliases: []string{"notifications", "notify"},
	Short:   "Manage notifications",
	Long:    `Commands for viewing notifications on your Unraid server.`,
}

var notificationListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		query := api.NotificationsQuery
		if showAll {
			query = api.AllNotificationsQuery
		}

		var resp api.NotificationsResponse
		if err := apiClient.Query(ctx, query, nil, &resp); err != nil {
			return fmt.Errorf("failed to list notifications: %w", err)
		}

		notifications := resp.Notifications.List

		if out.IsJSON() {
			return out.JSON(notifications)
		}

		if len(notifications) == 0 {
			out.Println("No notifications")
			return nil
		}

		if !showAll {
			out.Print("Unread notifications: %d\n\n", resp.Notifications.Overview.Unread.Total)
		}

		headers := []string{"IMPORTANCE", "SUBJECT", "TIMESTAMP"}
		var rows [][]string
		for _, n := range notifications {
			// Parse and format timestamp
			ts := n.Timestamp
			if t, err := time.Parse(time.RFC3339, n.Timestamp); err == nil {
				ts = t.Local().Format("2006-01-02 15:04")
			}
			rows = append(rows, []string{
				n.Importance,
				truncate(n.Subject, 50),
				ts,
			})
		}
		out.Table(headers, rows)

		return nil
	},
}

func init() {
	notificationCmd.AddCommand(notificationListCmd)
	notificationListCmd.Flags().BoolVarP(&showAll, "all", "a", false, "show all notifications (not just unread)")
}
