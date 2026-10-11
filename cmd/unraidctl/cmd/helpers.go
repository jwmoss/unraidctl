package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"

	"github.com/jwmoss/unraidctl/pkg/client"
	"github.com/spf13/cobra"
)

var (
	watchMode  bool
	watchCount int
)

func addWatchFlags(cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		cmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "stream live updates until interrupted")
		cmd.Flags().IntVar(&watchCount, "count", 0, "with --watch, stop after this many updates (0 = no limit)")
	}
}

// watch streams a subscription and renders each event until Ctrl-C or --count events.
func watch[T any](subscription string, variables map[string]interface{}, render func(T) error) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	seen := 0
	return apiClient.Subscribe(ctx, subscription, variables, func(data json.RawMessage) error {
		var event T
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to parse update: %w", err)
		}
		if err := render(event); err != nil {
			return err
		}
		seen++
		if watchCount > 0 && seen >= watchCount {
			return client.ErrStopSubscription
		}
		return nil
	})
}

// requireYes guards changes that are hard to undo.
func requireYes(cmd *cobra.Command, action string) error {
	if yes, _ := cmd.Flags().GetBool("yes"); !yes {
		return fmt.Errorf("%s; rerun with --yes to confirm", action)
	}
	return nil
}

// printJSONOr writes value as JSON in --json mode; otherwise it calls render.
func printJSONOr(value interface{}, render func()) error {
	if out.IsJSON() {
		return out.JSON(value)
	}
	render()
	return nil
}

func queryWithTimeout(query string, variables map[string]interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return apiClient.Query(ctx, query, variables, result)
}

func formatSizeBytes(bytes int64) string {
	if bytes == 0 {
		return "-"
	}

	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1f TiB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1f GiB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MiB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KiB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// reportResult prints the outcome of a mutation that returns a Boolean.
func reportResult(ok bool, success, failure string) error {
	if out.IsJSON() {
		return out.JSON(map[string]bool{"ok": ok})
	}
	if ok {
		out.Success(success)
	} else {
		out.Warn(failure)
	}
	return nil
}

// renderFields prints a flat object as a sorted FIELD/VALUE table.
func renderFields(fields map[string]interface{}) {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := make([][]string, 0, len(keys))
	for _, key := range keys {
		value := "-"
		if fields[key] != nil {
			value = fmt.Sprint(fields[key])
		}
		rows = append(rows, []string{key, value})
	}
	out.Table([]string{"FIELD", "VALUE"}, rows)
}
