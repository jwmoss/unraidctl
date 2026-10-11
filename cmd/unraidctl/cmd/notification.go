package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var showAll bool

const notificationPageSize = 100

var notificationCmd = &cobra.Command{
	Use:     "notification",
	Aliases: []string{"notifications", "notify"},
	Short:   "View and manage notifications",
}

var (
	notificationAllFlag    bool
	notificationImportance string
	notificationArchived   bool
	notificationTitle      string
	notificationSubject    string
	notificationDesc       string
	notificationLink       string
	notificationUnique     bool
)

var notificationArchiveCmd = notificationMoveCommand("archive", api.NotificationArchiveManyMutation, api.NotificationArchiveAllMutation)
var notificationUnarchiveCmd = notificationMoveCommand("unarchive", api.NotificationUnarchiveManyMutation, api.NotificationUnarchiveAllMutation)

// notificationMoveCommand moves the given IDs, or every notification (optionally of one importance) with --all.
func notificationMoveCommand(action, manyMutation, allMutation string) *cobra.Command {
	return &cobra.Command{
		Use:   action + " [notification-id...]",
		Args:  cobra.ArbitraryArgs,
		Short: strings.ToUpper(action[:1]) + action[1:] + " notifications by ID, or all with --all",
		RunE: func(cmd *cobra.Command, args []string) error {
			if (len(args) > 0) == notificationAllFlag {
				return fmt.Errorf("give notification IDs or --all, not both")
			}
			mutation, variables := manyMutation, map[string]interface{}{"ids": args}
			if notificationAllFlag {
				mutation, variables = allMutation, map[string]interface{}{"importance": importanceOrNil()}
			}
			return runNotificationOverviewMutation(mutation, variables, action)
		},
	}
}

var notificationDeleteCmd = &cobra.Command{
	Use: "delete <notification-id>", Short: "Delete a notification permanently", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireYes(cmd, "deleting a notification cannot be undone"); err != nil {
			return err
		}
		kind := "UNREAD"
		if notificationArchived {
			kind = "ARCHIVE"
		}
		return runNotificationOverviewMutation(api.NotificationDeleteMutation, map[string]interface{}{"id": args[0], "type": kind}, "delete")
	},
}

var notificationDeleteArchivedCmd = &cobra.Command{
	Use: "delete-archived", Short: "Delete every archived notification permanently", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireYes(cmd, "deleting archived notifications cannot be undone"); err != nil {
			return err
		}
		return runNotificationOverviewMutation(api.NotificationDeleteArchivedMutation, nil, "delete-archived")
	},
}

var notificationRecalculateCmd = &cobra.Command{
	Use: "recalculate", Short: "Recount notifications from disk", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNotificationOverviewMutation(api.NotificationRecalculateMutation, nil, "recalculate")
	},
}

var notificationCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a notification", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		input := map[string]interface{}{
			"title": notificationTitle, "subject": notificationSubject, "description": notificationDesc,
			"importance": strings.ToUpper(notificationImportance),
		}
		if notificationLink != "" {
			input["link"] = notificationLink
		}
		mutation, field := api.NotificationCreateMutation, "createNotification"
		if notificationUnique {
			mutation, field = api.NotificationNotifyIfUniqueMutation, "notifyIfUnique"
		}
		var resp map[string]*api.Notification
		if err := queryWithTimeout(mutation, map[string]interface{}{"input": input}, &resp); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
		created := resp[field]
		if created == nil {
			return printJSONOr(nil, func() { out.Println("An identical unread notification already exists") })
		}
		return printJSONOr(created, func() { renderNotifications([]api.Notification{*created}) })
	},
}

var notificationWatchCmd = &cobra.Command{
	Use: "watch", Short: "Stream new notifications", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return watch(api.NotificationSubscription, nil, func(r struct {
			Notification api.Notification `json:"notificationAdded"`
		}) error {
			return printJSONOr(r.Notification, func() { renderNotifications([]api.Notification{r.Notification}) })
		})
	},
}

func importanceOrNil() interface{} {
	if notificationImportance == "" {
		return nil
	}
	return strings.ToUpper(notificationImportance)
}

func runNotificationOverviewMutation(mutation string, variables map[string]interface{}, action string) error {
	var resp map[string]api.NotificationOverview
	if err := queryWithTimeout(mutation, variables, &resp); err != nil {
		return fmt.Errorf("failed to %s notifications: %w", action, err)
	}
	for _, overview := range resp {
		return printJSONOr(overview, func() {
			out.Success(fmt.Sprintf("Notification %s complete", action))
			out.Print("Unread: %d, archived: %d\n", overview.Unread.Total, overview.Archive.Total)
		})
	}
	return fmt.Errorf("API returned no notification overview")
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
		rows = append(rows, []string{n.ID, n.Type, n.Importance, n.Subject, ts})
	}
	out.Table([]string{"ID", "TYPE", "IMPORTANCE", "SUBJECT", "TIMESTAMP"}, rows)
}

func init() {
	notificationCmd.AddCommand(notificationListCmd, notificationAlertsCmd, notificationArchiveCmd, notificationUnarchiveCmd,
		notificationDeleteCmd, notificationDeleteArchivedCmd, notificationRecalculateCmd, notificationCreateCmd, notificationWatchCmd)
	notificationListCmd.Flags().BoolVarP(&showAll, "all", "a", false, "include archived notifications")
	for _, cmd := range []*cobra.Command{notificationArchiveCmd, notificationUnarchiveCmd} {
		cmd.Flags().BoolVar(&notificationAllFlag, "all", false, "apply to every notification")
		cmd.Flags().StringVar(&notificationImportance, "importance", "", "with --all, only INFO, WARNING, or ALERT")
	}
	notificationDeleteCmd.Flags().BoolVar(&notificationArchived, "archived", false, "the notification is archived, not unread")
	notificationDeleteCmd.Flags().Bool("yes", false, "confirm the deletion")
	notificationDeleteArchivedCmd.Flags().Bool("yes", false, "confirm the deletion")
	notificationCreateCmd.Flags().StringVar(&notificationTitle, "title", "", "notification title")
	notificationCreateCmd.Flags().StringVar(&notificationSubject, "subject", "", "notification subject")
	notificationCreateCmd.Flags().StringVar(&notificationDesc, "description", "", "notification description")
	notificationCreateCmd.Flags().StringVar(&notificationImportance, "importance", "INFO", "INFO, WARNING, or ALERT")
	notificationCreateCmd.Flags().StringVar(&notificationLink, "link", "", "optional link")
	notificationCreateCmd.Flags().BoolVar(&notificationUnique, "unique", false, "skip when an identical unread notification exists")
	for _, flag := range []string{"title", "subject", "description"} {
		_ = notificationCreateCmd.MarkFlagRequired(flag)
	}
	notificationWatchCmd.Flags().IntVar(&watchCount, "count", 0, "stop after this many notifications (0 = no limit)")
}
