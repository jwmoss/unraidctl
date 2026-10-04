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
This creates or updates ~/.config/unraidctl/config.yaml, or the path selected with --config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		path := cfgFile
		if path == "" {
			path = config.DefaultConfigPath()
		}
		existingCfg, err := config.Load(path)
		if err != nil {
			return err
		}
		if serverURL != "" {
			existingCfg.Server = serverURL
		}
		if apiKey != "" {
			existingCfg.APIKey = apiKey
		}

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

		if err := newCfg.Validate(); err != nil {
			return err
		}
		if err := config.Save(newCfg, path); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println()
		fmt.Printf("Configuration saved to %s\n", path)
		fmt.Println()
		fmt.Println("Test your connection with:")
		fmt.Println("  unraidctl info")

		return nil
	},
}
