package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Inspect services, network interfaces, plugins, and system time",
}

var systemServicesCmd = &cobra.Command{
	Use: "services", Short: "List API services and their state", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp struct {
			Services []struct {
				Name    string `json:"name"`
				Online  *bool  `json:"online"`
				Version string `json:"version"`
				Uptime  *struct {
					Timestamp string `json:"timestamp"`
				} `json:"uptime"`
			} `json:"services"`
		}
		if err := queryWithTimeout(api.ServicesQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list services: %w", err)
		}
		return printJSONOr(resp.Services, func() {
			rows := make([][]string, 0, len(resp.Services))
			for _, s := range resp.Services {
				since := "-"
				if s.Uptime != nil {
					since = s.Uptime.Timestamp
				}
				rows = append(rows, []string{s.Name, fmt.Sprint(s.Online != nil && *s.Online), s.Version, since})
			}
			out.Table([]string{"NAME", "ONLINE", "VERSION", "SINCE"}, rows)
		})
	},
}

var systemServerCmd = &cobra.Command{
	Use: "server", Short: "Show server identity and addresses", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp struct {
			Server map[string]interface{} `json:"server"`
		}
		if err := queryWithTimeout(api.ServerQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get server: %w", err)
		}
		return printJSONOr(resp.Server, func() { renderFields(resp.Server) })
	},
}

var systemInterfacesCmd = &cobra.Command{
	Use: "interfaces", Short: "List network interface configuration", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp struct {
			Interfaces json.RawMessage `json:"networkInterfaces"`
		}
		if err := queryWithTimeout(api.NetworkInterfacesQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list network interfaces: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.Interfaces)
		}
		var interfaces []struct {
			Name       string `json:"name"`
			MACAddress string `json:"macAddress"`
			Operstate  string `json:"operstate"`
			Speed      *int   `json:"speed"`
			IPAddress  string `json:"ipAddress"`
			Netmask    string `json:"netmask"`
			Gateway    string `json:"gateway"`
			UseDHCP    *bool  `json:"useDhcp"`
		}
		if err := json.Unmarshal(resp.Interfaces, &interfaces); err != nil {
			return fmt.Errorf("failed to parse network interfaces: %w", err)
		}
		rows := make([][]string, 0, len(interfaces))
		for _, i := range interfaces {
			rows = append(rows, []string{i.Name, i.Operstate, optionalInt(i.Speed, " Mb/s"), i.IPAddress, i.Netmask, i.Gateway, fmt.Sprint(i.UseDHCP != nil && *i.UseDHCP), i.MACAddress})
		}
		out.Table([]string{"NAME", "STATE", "SPEED", "IPV4", "NETMASK", "GATEWAY", "DHCP", "MAC"}, rows)
		return nil
	},
}

var systemPluginsCmd = &cobra.Command{
	Use: "plugins", Short: "List installed Unraid plugins and API plugins", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp struct {
			APIPlugins []struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"plugins"`
			Installed []string `json:"installedUnraidPlugins"`
		}
		if err := queryWithTimeout(api.PluginsQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list plugins: %w", err)
		}
		return printJSONOr(resp, func() {
			rows := make([][]string, 0, len(resp.Installed)+len(resp.APIPlugins))
			for _, name := range resp.Installed {
				rows = append(rows, []string{"unraid", name, "-"})
			}
			for _, p := range resp.APIPlugins {
				rows = append(rows, []string{"api", p.Name, p.Version})
			}
			out.Table([]string{"KIND", "NAME", "VERSION"}, rows)
		})
	},
}

type systemTime struct {
	CurrentTime string   `json:"currentTime"`
	TimeZone    string   `json:"timeZone"`
	UseNTP      bool     `json:"useNtp"`
	NTPServers  []string `json:"ntpServers"`
}

func renderSystemTime(t systemTime) error {
	return printJSONOr(t, func() {
		servers := strings.Join(strings.Fields(strings.Join(t.NTPServers, " ")), ", ")
		out.Table([]string{"FIELD", "VALUE"}, [][]string{{"Current time", t.CurrentTime}, {"Time zone", t.TimeZone}, {"NTP", fmt.Sprint(t.UseNTP)}, {"NTP servers", servers}})
	})
}

var systemTimeCmd = &cobra.Command{
	Use: "time", Short: "Show system time, time zone, and NTP settings", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp struct {
			SystemTime systemTime `json:"systemTime"`
		}
		if err := queryWithTimeout(api.SystemTimeQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to get system time: %w", err)
		}
		return renderSystemTime(resp.SystemTime)
	},
}

var systemTimezonesCmd = &cobra.Command{
	Use: "timezones", Short: "List time zones that set-time accepts", Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp struct {
			Options []struct {
				Value string `json:"value"`
				Label string `json:"label"`
			} `json:"timeZoneOptions"`
		}
		if err := queryWithTimeout(api.TimeZoneOptionsQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list time zones: %w", err)
		}
		return printJSONOr(resp.Options, func() {
			rows := make([][]string, 0, len(resp.Options))
			for _, o := range resp.Options {
				rows = append(rows, []string{o.Value, o.Label})
			}
			out.Table([]string{"VALUE", "LABEL"}, rows)
		})
	},
}

var systemSetTimeCmd = &cobra.Command{
	Use:   "set-time",
	Short: "Change the time zone, NTP settings, or clock",
	Long:  "Change only the settings given as flags. --manual takes \"YYYY-MM-DD HH:mm:ss\" and applies with --ntp=false.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		input := map[string]interface{}{}
		flags := cmd.Flags()
		if flags.Changed("timezone") {
			input["timeZone"], _ = flags.GetString("timezone")
		}
		if flags.Changed("ntp") {
			input["useNtp"], _ = flags.GetBool("ntp")
		}
		if flags.Changed("ntp-server") {
			input["ntpServers"], _ = flags.GetStringSlice("ntp-server")
		}
		if flags.Changed("manual") {
			input["manualDateTime"], _ = flags.GetString("manual")
		}
		if len(input) == 0 {
			return fmt.Errorf("set at least one of --timezone, --ntp, --ntp-server, or --manual")
		}
		if err := requireYes(cmd, "this changes the server clock settings"); err != nil {
			return err
		}
		var resp struct {
			SystemTime systemTime `json:"updateSystemTime"`
		}
		if err := queryWithTimeout(api.UpdateSystemTimeMutation, map[string]interface{}{"input": input}, &resp); err != nil {
			return fmt.Errorf("failed to update system time: %w", err)
		}
		out.Success("System time settings updated")
		return renderSystemTime(resp.SystemTime)
	},
}

var flashCmd = &cobra.Command{Use: "flash", Short: "Back up the USB flash boot device"}

var flashBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Start an rclone backup of the flash device",
	Long:  "Start an rclone backup of the flash device to a configured rclone remote. --options takes a JSON object of rclone options.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		input := map[string]interface{}{}
		for flag, key := range map[string]string{"remote": "remoteName", "source": "sourcePath", "destination": "destinationPath"} {
			input[key], _ = flags.GetString(flag)
		}
		if options, _ := flags.GetString("options"); options != "" {
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(options), &parsed); err != nil {
				return fmt.Errorf("--options must be a JSON object: %w", err)
			}
			input["options"] = parsed
		}
		var resp struct {
			Status struct {
				Status string  `json:"status"`
				JobID  *string `json:"jobId"`
			} `json:"initiateFlashBackup"`
		}
		if err := queryWithTimeout(api.FlashBackupMutation, map[string]interface{}{"input": input}, &resp); err != nil {
			return fmt.Errorf("failed to start flash backup: %w", err)
		}
		return printJSONOr(resp.Status, func() {
			job := "-"
			if resp.Status.JobID != nil {
				job = *resp.Status.JobID
			}
			out.Success(fmt.Sprintf("Flash backup %s (job %s)", resp.Status.Status, job))
		})
	},
}

func init() {
	systemCmd.AddCommand(systemServicesCmd, systemServerCmd, systemInterfacesCmd, systemPluginsCmd, systemTimeCmd, systemTimezonesCmd, systemSetTimeCmd)
	flashCmd.AddCommand(flashBackupCmd)
	rootCmd.AddCommand(systemCmd, flashCmd)

	systemSetTimeCmd.Flags().String("timezone", "", "time zone, for example America/New_York")
	systemSetTimeCmd.Flags().Bool("ntp", true, "use NTP")
	systemSetTimeCmd.Flags().StringSlice("ntp-server", nil, "NTP server (repeat up to four times)")
	systemSetTimeCmd.Flags().String("manual", "", "manual date and time, YYYY-MM-DD HH:mm:ss")
	systemSetTimeCmd.Flags().Bool("yes", false, "confirm the change")

	flashBackupCmd.Flags().String("remote", "", "rclone remote name")
	flashBackupCmd.Flags().String("source", "/boot", "source path")
	flashBackupCmd.Flags().String("destination", "", "destination path on the remote")
	flashBackupCmd.Flags().String("options", "", "rclone options as a JSON object")
	for _, flag := range []string{"remote", "destination"} {
		_ = flashBackupCmd.MarkFlagRequired(flag)
	}
}
