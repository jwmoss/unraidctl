package cmd

import (
	"fmt"
	"strings"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var ssoCmd = &cobra.Command{
	Use:   "sso",
	Short: "Inspect SSO and OIDC configuration",
}

var ssoStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether SSO is enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.OIDCProvidersResponse
		if err := queryWithTimeout(api.OIDCProvidersQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to read SSO status: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(map[string]interface{}{
				"enabled":         resp.IsSSOEnabled,
				"providers":       len(resp.OIDCProviders),
				"publicProviders": len(resp.PublicOIDCProviders),
			})
		}
		out.Print("SSO enabled: %t\n", resp.IsSSOEnabled)
		out.Print("OIDC providers: %d\n", len(resp.OIDCProviders))
		out.Print("Public providers: %d\n", len(resp.PublicOIDCProviders))
		return nil
	},
}

var ssoProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List configured OIDC providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.OIDCProvidersResponse
		if err := queryWithTimeout(api.OIDCProvidersQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list OIDC providers: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.OIDCProviders)
		}
		renderOIDCProviders(resp.OIDCProviders)
		return nil
	},
}

var ssoPublicProvidersCmd = &cobra.Command{
	Use:   "public-providers",
	Short: "List public OIDC login providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.OIDCProvidersResponse
		if err := queryWithTimeout(api.OIDCProvidersQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list public OIDC providers: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.PublicOIDCProviders)
		}
		rows := make([][]string, 0, len(resp.PublicOIDCProviders))
		for _, provider := range resp.PublicOIDCProviders {
			rows = append(rows, []string{provider.ID, provider.Name, provider.ButtonText, provider.ButtonVariant})
		}
		out.Table([]string{"ID", "NAME", "BUTTON", "VARIANT"}, rows)
		return nil
	},
}

var ssoConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show OIDC configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.OIDCConfigurationResponse
		if err := queryWithTimeout(api.OIDCConfigurationQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to read OIDC configuration: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.OIDCConfiguration)
		}
		out.Print("Default allowed origins: %s\n\n", strings.Join(resp.OIDCConfiguration.DefaultAllowedOrigins, ", "))
		renderOIDCProviders(resp.OIDCConfiguration.Providers)
		return nil
	},
}

var ssoValidateTokenCmd = &cobra.Command{
	Use:   "validate-token <token>",
	Short: "Validate an OIDC session token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.ValidateOIDCSessionResponse
		if err := queryWithTimeout(api.ValidateOIDCSessionQuery, map[string]interface{}{"token": args[0]}, &resp); err != nil {
			return fmt.Errorf("failed to validate OIDC session token: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.ValidateOIDCSession)
		}
		out.Print("Valid: %t\n", resp.ValidateOIDCSession.Valid)
		if resp.ValidateOIDCSession.Username != "" {
			out.Print("Username: %s\n", resp.ValidateOIDCSession.Username)
		}
		return nil
	},
}

func renderOIDCProviders(providers []api.OIDCProvider) {
	rows := make([][]string, 0, len(providers))
	for _, provider := range providers {
		rows = append(rows, []string{
			provider.ID,
			provider.Name,
			provider.ClientID,
			provider.Issuer,
			strings.Join(provider.Scopes, ","),
		})
	}
	out.Table([]string{"ID", "NAME", "CLIENT", "ISSUER", "SCOPES"}, rows)
}

func init() {
	ssoCmd.AddCommand(ssoStatusCmd)
	ssoCmd.AddCommand(ssoProvidersCmd)
	ssoCmd.AddCommand(ssoPublicProvidersCmd)
	ssoCmd.AddCommand(ssoConfigCmd)
	ssoCmd.AddCommand(ssoValidateTokenCmd)
}
