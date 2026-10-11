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
		if watchMode {
			return watch(api.NetworkSubscription, nil, func(r struct {
				Network []api.NetworkMetrics `json:"systemMetricsNetwork"`
			}) error {
				return renderNetworkMetrics(r.Network)
			})
		}
		var resp api.MetricsResponse
		if err := queryWithTimeout(api.NetworkMetricsQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get network metrics: %w", err)
		}
		return renderNetworkMetrics(resp.Metrics.Network)
	},
}

func renderNetworkMetrics(metrics []api.NetworkMetrics) error {
	if out.IsJSON() {
		return out.JSON(metrics)
	}
	if len(metrics) == 0 {
		out.Println("No network interfaces found")
		return nil
	}
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
	return nil
}

func formatRate(bytesPerSecond float64) string {
	if bytesPerSecond == 0 {
		return "-"
	}
	return formatSizeBytes(int64(bytesPerSecond)) + "/s"
}

func init() {
	metricsCmd.AddCommand(networkMetricsCmd, cpuMetricsCmd, memoryMetricsCmd, temperatureMetricsCmd, cpuTelemetryCmd)
	addWatchFlags(networkMetricsCmd, cpuMetricsCmd, memoryMetricsCmd, temperatureMetricsCmd)
	cpuTelemetryCmd.Flags().IntVar(&watchCount, "count", 0, "stop after this many updates (0 = no limit)")
}

var cpuMetricsCmd = &cobra.Command{
	Use: "cpu", Short: "View CPU utilization", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		render := func(cpu *api.CPUUtilization) error {
			if cpu == nil {
				return fmt.Errorf("CPU metrics are unavailable")
			}
			return printJSONOr(cpu, func() { out.Print("CPU utilization: %.1f%%\n", cpu.PercentTotal) })
		}
		if watchMode {
			return watch(api.CPUSubscription, nil, func(r struct {
				CPU *api.CPUUtilization `json:"systemMetricsCpu"`
			}) error {
				return render(r.CPU)
			})
		}
		var resp api.MetricsResponse
		if err := queryWithTimeout(api.CPUMetricsQuery, nil, &resp); err != nil {
			return err
		}
		return render(resp.Metrics.CPU)
	},
}
var memoryMetricsCmd = &cobra.Command{
	Use: "memory", Short: "View active and available memory", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		render := func(memory *api.MemoryUtilization) error {
			if memory == nil {
				return fmt.Errorf("memory metrics are unavailable")
			}
			return printJSONOr(memory, func() {
				out.Table([]string{"TOTAL", "ACTIVE", "AVAILABLE", "USED"}, [][]string{{formatSizeBytes(memory.Total), formatSizeBytes(memory.Active), formatSizeBytes(memory.Available), fmt.Sprintf("%.1f%%", memory.PercentTotal)}})
			})
		}
		if watchMode {
			return watch(api.MemorySubscription, nil, func(r struct {
				Memory *api.MemoryUtilization `json:"systemMetricsMemory"`
			}) error {
				return render(r.Memory)
			})
		}
		var resp api.MetricsResponse
		if err := queryWithTimeout(api.MemoryMetricsQuery, nil, &resp); err != nil {
			return err
		}
		return render(resp.Metrics.Memory)
	},
}
var temperatureMetricsCmd = &cobra.Command{
	Use: "temperature", Short: "View temperature sensors and their status", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if watchMode {
			return watch(api.TemperatureSubscription, nil, func(r struct {
				Temperature *api.TemperatureMetrics `json:"systemMetricsTemperature"`
			}) error {
				return renderTemperature(r.Temperature)
			})
		}
		var resp api.MetricsResponse
		if err := queryWithTimeout(api.TemperatureMetricsQuery, nil, &resp); err != nil {
			return err
		}
		return renderTemperature(resp.Metrics.Temperature)
	},
}

func renderTemperature(temperature *api.TemperatureMetrics) error {
	if temperature == nil {
		return fmt.Errorf("temperature metrics are unavailable")
	}
	return printJSONOr(temperature, func() {
		rows := make([][]string, 0, len(temperature.Sensors))
		for _, sensor := range temperature.Sensors {
			rows = append(rows, []string{sensor.Name, sensor.Type, fmt.Sprintf("%.1f %s", sensor.Current.Value, sensor.Current.Unit), sensor.Current.Status})
		}
		out.Table([]string{"SENSOR", "TYPE", "TEMPERATURE", "STATUS"}, rows)
	})
}

// cpuTelemetryCmd has no query form; the API publishes package power and temperature only as a subscription.
var cpuTelemetryCmd = &cobra.Command{
	Use: "cpu-telemetry", Short: "Stream CPU package power and temperature", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return watch(api.CPUTelemetrySubscription, nil, func(r struct {
			Telemetry api.CPUPackages `json:"systemMetricsCpuTelemetry"`
		}) error {
			t := r.Telemetry
			return printJSONOr(t, func() {
				out.Print("Total power: %.1f W | package power (W): %v | package temperature (C): %v\n", t.TotalPower, t.Power, t.Temp)
			})
		})
	},
}
