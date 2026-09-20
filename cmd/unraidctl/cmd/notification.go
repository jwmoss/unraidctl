package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var showAll bool

const notificationPageSize = 100

var notificationCmd = &cobra.Command{
	Use:     "notification",
	Aliases: []string{"notifications", "notify"},
	Short:   "View notifications",
}

var notificationListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List every unread notification, or include archived notifications with --all",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		notifications, err := listNotifications(showAll)
		if err != nil {
			return err
		}
		if out.IsJSON() {
			return out.JSON(notifications)
		}
		renderNotifications(notifications)
		return nil
	},
}

var notificationAlertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Show deduplicated unread warnings and alerts",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.NotificationsResponse
		if err := queryWithTimeout(api.NotificationAlertsQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get alerts: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.Notifications)
		}
		counts := resp.Notifications.Overview.Unread
		out.Print("Unread: %d alerts, %d warnings, %d informational\n", counts.Alert, counts.Warning, counts.Info)
		renderNotifications(resp.Notifications.WarningsAndAlerts)
		return nil
	},
}

func listNotifications(all bool) ([]api.Notification, error) {
	kinds := []string{"UNREAD"}
	if all {
		kinds = append(kinds, "ARCHIVE")
	}
	notifications := make([]api.Notification, 0)
	for _, kind := range kinds {
		page, err := notificationPages(kind)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, page...)
	}
	sort.SliceStable(notifications, func(i, j int) bool { return notifications[i].Timestamp > notifications[j].Timestamp })
	return notifications, nil
}

func notificationPages(kind string) ([]api.Notification, error) {
	notifications := make([]api.Notification, 0)
	for offset := 0; ; offset += notificationPageSize {
		filter := map[string]interface{}{"type": kind, "offset": offset, "limit": notificationPageSize}
		var resp api.NotificationsResponse
		if err := queryWithTimeout(api.NotificationsQuery, map[string]interface{}{"filter": filter}, &resp); err != nil {
			return nil, fmt.Errorf("failed to list %s notifications at offset %d: %w", kind, offset, err)
		}
		notifications = append(notifications, resp.Notifications.List...)
		if len(resp.Notifications.List) < notificationPageSize {
			return notifications, nil
		}
	}
}

func renderNotifications(notifications []api.Notification) {
	if len(notifications) == 0 {
		out.Println("No notifications")
		return
	}
	out.Print("Notifications: %d\n\n", len(notifications))
	rows := make([][]string, 0, len(notifications))
	for _, n := range notifications {
		ts := n.Timestamp
		if stamp, err := time.Parse(time.RFC3339, ts); err == nil {
			ts = stamp.Local().Format("2006-01-02 15:04")
		}
		rows = append(rows, []string{n.Type, n.Importance, n.Subject, ts})
	}
	out.Table([]string{"TYPE", "IMPORTANCE", "SUBJECT", "TIMESTAMP"}, rows)
}

func init() {
	notificationCmd.AddCommand(notificationListCmd, notificationAlertsCmd)
	notificationListCmd.Flags().BoolVarP(&showAll, "all", "a", false, "include archived notifications")
}
