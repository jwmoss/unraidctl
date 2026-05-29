package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:     "settings",
	Aliases: []string{"setting"},
	Short:   "Inspect and update Unraid API settings",
}

var (
	settingsUpdateFile string
	settingsUpdateData string
	settingsShowValues bool
)

var settingsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show API, SSO, and unified settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.SettingsResponse
		if err := queryWithTimeout(api.SettingsQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to read settings: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp)
		}
		renderSettings(resp)
		if settingsShowValues && len(resp.Settings.Unified.Values) > 0 {
			out.Println("")
			out.Println("Unified Values")
			out.Println(string(resp.Settings.Unified.Values))
		}
		return nil
	},
}

var settingsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update settings from a JSON object",
	RunE: func(cmd *cobra.Command, args []string) error {
		if (settingsUpdateFile == "") == (settingsUpdateData == "") {
			return fmt.Errorf("set exactly one of --file or --data")
		}

		payloadBytes := []byte(settingsUpdateData)
		if settingsUpdateFile != "" {
			bytes, err := os.ReadFile(settingsUpdateFile)
			if err != nil {
				return fmt.Errorf("failed to read settings file: %w", err)
			}
			payloadBytes = bytes
		}

		var payload interface{}
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return fmt.Errorf("settings payload must be valid JSON: %w", err)
		}

		var resp api.UpdateSettingsResponse
		if err := queryWithTimeout(api.UpdateSettingsMutation, map[string]interface{}{"input": payload}, &resp); err != nil {
			return fmt.Errorf("failed to update settings: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.UpdateSettings)
		}
		out.Success("Settings updated")
		out.Print("Restart required: %t\n", resp.UpdateSettings.RestartRequired)
		if len(resp.UpdateSettings.Warnings) > 0 {
			out.Println("Warnings:")
			for _, warning := range resp.UpdateSettings.Warnings {
				out.Print("  - %s\n", warning)
			}
		}
		return nil
	},
}

func renderSettings(resp api.SettingsResponse) {
	apiCfg := resp.Settings.API
	sandbox := "-"
	if apiCfg.Sandbox != nil {
		sandbox = fmt.Sprintf("%t", *apiCfg.Sandbox)
	}
	rows := [][]string{
		{"API version", apiCfg.Version},
		{"Sandbox", sandbox},
		{"SSO enabled", fmt.Sprintf("%t", resp.IsSSOEnabled)},
		{"Extra origins", strings.Join(apiCfg.ExtraOrigins, ", ")},
		{"SSO subject IDs", strings.Join(apiCfg.SSOSubIDs, ", ")},
		{"Plugins", strings.Join(apiCfg.Plugins, ", ")},
		{"OIDC providers", fmt.Sprintf("%d", len(resp.Settings.SSO.OIDCProviders))},
	}
	out.Table([]string{"SETTING", "VALUE"}, rows)
}

func init() {
	settingsCmd.AddCommand(settingsShowCmd)
	settingsCmd.AddCommand(settingsUpdateCmd)

	settingsShowCmd.Flags().BoolVar(&settingsShowValues, "values", false, "print raw unified settings values")
	settingsUpdateCmd.Flags().StringVar(&settingsUpdateFile, "file", "", "path to JSON settings payload")
	settingsUpdateCmd.Flags().StringVar(&settingsUpdateData, "data", "", "inline JSON settings payload")
}
