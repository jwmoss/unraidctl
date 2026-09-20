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
	metricsCmd.AddCommand(networkMetricsCmd, cpuMetricsCmd, memoryMetricsCmd, temperatureMetricsCmd)
}

var cpuMetricsCmd = &cobra.Command{
	Use: "cpu", Short: "View CPU utilization", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.MetricsResponse
		if err := queryWithTimeout(api.CPUMetricsQuery, nil, &resp); err != nil {
			return err
		}
		if resp.Metrics.CPU == nil {
			return fmt.Errorf("CPU metrics are unavailable")
		}
		if out.IsJSON() {
			return out.JSON(resp.Metrics.CPU)
		}
		out.Print("CPU utilization: %.1f%%\n", resp.Metrics.CPU.PercentTotal)
		return nil
	},
}
var memoryMetricsCmd = &cobra.Command{
	Use: "memory", Short: "View active and available memory", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.MetricsResponse
		if err := queryWithTimeout(api.MemoryMetricsQuery, nil, &resp); err != nil {
			return err
		}
		if resp.Metrics.Memory == nil {
			return fmt.Errorf("memory metrics are unavailable")
		}
		if out.IsJSON() {
			return out.JSON(resp.Metrics.Memory)
		}
		memory := resp.Metrics.Memory
		out.Table([]string{"TOTAL", "ACTIVE", "AVAILABLE", "USED"}, [][]string{{formatSizeBytes(memory.Total), formatSizeBytes(memory.Active), formatSizeBytes(memory.Available), fmt.Sprintf("%.1f%%", memory.PercentTotal)}})
		return nil
	},
}
var temperatureMetricsCmd = &cobra.Command{
	Use: "temperature", Short: "View temperature sensors and their status", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.MetricsResponse
		if err := queryWithTimeout(api.TemperatureMetricsQuery, nil, &resp); err != nil {
			return err
		}
		if resp.Metrics.Temperature == nil {
			return fmt.Errorf("temperature metrics are unavailable")
		}
		if out.IsJSON() {
			return out.JSON(resp.Metrics.Temperature)
		}
		rows := make([][]string, 0, len(resp.Metrics.Temperature.Sensors))
		for _, sensor := range resp.Metrics.Temperature.Sensors {
			rows = append(rows, []string{sensor.Name, sensor.Type, fmt.Sprintf("%.1f %s", sensor.Current.Value, sensor.Current.Unit), sensor.Current.Status})
		}
		out.Table([]string{"SENSOR", "TYPE", "TEMPERATURE", "STATUS"}, rows)
		return nil
	},
}
