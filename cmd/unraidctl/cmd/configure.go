package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jwmoss/unraidctl/internal/config"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure unraidctl settings",
	Long: `Interactively configure unraidctl with your Unraid server details.
This will create or update the config file at ~/.config/unraidctl/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		// Load existing config if present
		existingCfg, _ := config.Load("")

		fmt.Println("unraidctl Configuration")
		fmt.Println("=======================")
		fmt.Println()

		// Server URL
		defaultServer := existingCfg.Server
		if defaultServer == "" {
			defaultServer = "http://192.168.1.100"
		}
		fmt.Printf("Unraid server URL [%s]: ", defaultServer)
		serverInput, _ := reader.ReadString('\n')
		serverInput = strings.TrimSpace(serverInput)
		if serverInput == "" {
			serverInput = defaultServer
		}

		// API Key
		fmt.Print("API key: ")
		apiKeyInput, _ := reader.ReadString('\n')
		apiKeyInput = strings.TrimSpace(apiKeyInput)
		if apiKeyInput == "" && existingCfg.APIKey != "" {
			apiKeyInput = existingCfg.APIKey
			fmt.Println("(keeping existing API key)")
		}

		newCfg := &config.Config{
			Server: serverInput,
			APIKey: apiKeyInput,
		}

		if err := config.Save(newCfg, ""); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println()
		fmt.Printf("Configuration saved to %s\n", config.DefaultConfigPath())
		fmt.Println()
		fmt.Println("Test your connection with:")
		fmt.Println("  unraidctl info")

		return nil
	},
}
