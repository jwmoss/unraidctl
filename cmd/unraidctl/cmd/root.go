package cmd

import (
	"fmt"
	"os"

	"github.com/jwmoss/unraidctl/internal/config"
	"github.com/jwmoss/unraidctl/internal/output"
	"github.com/jwmoss/unraidctl/pkg/client"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	serverURL  string
	apiKey     string
	jsonOutput bool
	quiet      bool
	noColor    bool

	cfg       *config.Config
	apiClient *client.Client
	out       *output.Formatter
)

var rootCmd = &cobra.Command{
	Use:   "unraidctl",
	Short: "CLI tool to interact with the Unraid API",
	Long: `unraidctl is a command-line interface for managing your Unraid server
via the Unraid API. It provides commands for system information,
array management, Docker containers, VMs, shares, and notifications.

Configuration can be provided via:
  - Command-line flags (highest priority)
  - Environment variables (UNRAID_SERVER, UNRAID_API_KEY)
  - Config file (~/.config/unraidctl/config.yaml)`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config validation for help and version
		if cmd.Name() == "help" || cmd.Name() == "version" || cmd.Name() == "configure" {
			return nil
		}

		out = output.New(jsonOutput, quiet, noColor)

		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return err
		}

		// Override with flags
		if serverURL != "" {
			cfg.Server = serverURL
		}
		if apiKey != "" {
			cfg.APIKey = apiKey
		}

		// Validate config for commands that need API access
		if err := cfg.Validate(); err != nil {
			return err
		}

		apiClient = client.New(cfg.Server, cfg.APIKey)
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/unraidctl/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "Unraid server URL (env: UNRAID_SERVER)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key (env: UNRAID_API_KEY)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(arrayCmd)
	rootCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(vmCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(notificationCmd)
	rootCmd.AddCommand(configureCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("unraidctl version 0.1.0")
	},
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	// Could add more specific error type checking here
	return 1
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
