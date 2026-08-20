package cmd

import (
	"fmt"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "View system metrics",
}

var networkMetricsCmd = &cobra.Command{
	Use:   "network",
	Short: "View network interface metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.MetricsResponse
		if err := queryWithTimeout(api.NetworkMetricsQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get network metrics: %w", err)
		}

		metrics := resp.Metrics.Network
		if out.IsJSON() {
			return out.JSON(metrics)
		}
		if len(metrics) == 0 {
			out.Println("No network interfaces found")
			return nil
		}

		renderNetworkMetrics(metrics)
		return nil
	},
}

func renderNetworkMetrics(metrics []api.NetworkMetrics) {
	headers := []string{
		"INTERFACE", "STATE", "RX/S", "TX/S", "RECEIVED", "SENT", "ERRORS", "DROPPED", "UTILIZATION",
	}
	rows := make([][]string, 0, len(metrics))
	for _, metric := range metrics {
		utilization := "-"
		if metric.UtilizationPercent != nil {
			utilization = fmt.Sprintf("%.2f%%", *metric.UtilizationPercent)
		}
		rows = append(rows, []string{
			metric.Name,
			metric.Operstate,
			formatRate(metric.RxSec),
			formatRate(metric.TxSec),
			formatSizeBytes(metric.BytesReceived),
			formatSizeBytes(metric.BytesSent),
			fmt.Sprintf("%d/%d", metric.ReceiveErrors, metric.TransmitErrors),
			fmt.Sprintf("%d/%d", metric.ReceiveDropped, metric.TransmitDropped),
			utilization,
		})
	}
	out.Table(headers, rows)
}

func formatRate(bytesPerSecond float64) string {
	if bytesPerSecond == 0 {
		return "-"
	}
	return formatSizeBytes(int64(bytesPerSecond)) + "/s"
}

func init() {
	metricsCmd.AddCommand(networkMetricsCmd)
}
