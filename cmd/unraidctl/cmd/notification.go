package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var notificationCmd = &cobra.Command{
	Use:     "notification",
	Aliases: []string{"notifications", "notify"},
	Short:   "Manage notifications",
	Long:    `Commands for viewing and managing notifications on your Unraid server.`,
}

var notificationListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var resp api.NotificationsResponse
		if err := apiClient.Query(ctx, api.NotificationsQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list notifications: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(resp.Notifications)
		}

		if len(resp.Notifications) == 0 {
			out.Println("No notifications")
			return nil
		}

		headers := []string{"ID", "IMPORTANCE", "SUBJECT", "TIMESTAMP"}
		var rows [][]string
		for _, n := range resp.Notifications {
			idShort := n.ID
			if len(idShort) > 12 {
				idShort = idShort[:12]
			}
			rows = append(rows, []string{
				idShort,
				n.Importance,
				n.Subject,
				n.Timestamp,
			})
		}
		out.Table(headers, rows)

		return nil
	},
}

var notificationDismissCmd = &cobra.Command{
	Use:   "dismiss <id>",
	Short: "Dismiss a notification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		notifID := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		vars := map[string]interface{}{"id": notifID}
		if err := apiClient.Query(ctx, api.NotificationDismissMutation, vars, nil); err != nil {
			return fmt.Errorf("failed to dismiss notification: %w", err)
		}

		if out.IsJSON() {
			return out.JSON(map[string]string{"dismissed": notifID})
		}

		out.Success(fmt.Sprintf("Notification %s dismissed", notifID))
		return nil
	},
}

func init() {
	notificationCmd.AddCommand(notificationListCmd)
	notificationCmd.AddCommand(notificationDismissCmd)
}
