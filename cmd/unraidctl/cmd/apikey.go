package cmd

import (
	"fmt"
	"strings"

	"github.com/jwmoss/unraidctl/internal/api"
	"github.com/spf13/cobra"
)

var apiKeyCmd = &cobra.Command{
	Use:     "apikey",
	Aliases: []string{"api-key", "api-keys", "apikeys"},
	Short:   "Manage Unraid API keys",
}

var (
	apiKeyCreateName        string
	apiKeyCreateDescription string
	apiKeyCreateRoles       []string
	apiKeyCreatePermissions []string
	apiKeyCreateOverwrite   bool
	apiKeyUpdateName        string
	apiKeyUpdateDescription string
	apiKeyUpdateRoles       []string
	apiKeyUpdatePermissions []string
	apiKeyAddRole           string
	apiKeyRemoveRole        string
)

var apiKeyListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.APIKeysResponse
		if err := queryWithTimeout(api.APIKeysQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list API keys: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.APIKeys)
		}
		renderAPIKeys(resp.APIKeys)
		return nil
	},
}

var apiKeyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiKeyCreateName == "" {
			return fmt.Errorf("--name is required")
		}
		input, err := apiKeyInput(apiKeyCreateName, apiKeyCreateDescription, apiKeyCreateRoles, apiKeyCreatePermissions)
		if err != nil {
			return err
		}
		input["overwrite"] = apiKeyCreateOverwrite

		var resp api.APIKeyMutationResponse
		if err := queryWithTimeout(api.CreateAPIKeyMutation, map[string]interface{}{"input": input}, &resp); err != nil {
			return fmt.Errorf("failed to create API key: %w", err)
		}
		key := resp.APIKey.Create
		if out.IsJSON() {
			return out.JSON(key)
		}
		out.Success("API key created")
		out.Print("ID:   %s\n", key.ID)
		out.Print("Name: %s\n", key.Name)
		out.Print("Key:  %s\n", key.Key)
		return nil
	},
}

var apiKeyUpdateCmd = &cobra.Command{
	Use:   "update <api-key-id>",
	Short: "Update an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := apiKeyInput(apiKeyUpdateName, apiKeyUpdateDescription, apiKeyUpdateRoles, apiKeyUpdatePermissions)
		if err != nil {
			return err
		}
		input["id"] = args[0]

		var resp api.APIKeyMutationResponse
		if err := queryWithTimeout(api.UpdateAPIKeyMutation, map[string]interface{}{"input": input}, &resp); err != nil {
			return fmt.Errorf("failed to update API key: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.APIKey.Update)
		}
		out.Success("API key updated")
		renderAPIKeys([]api.APIKey{resp.APIKey.Update})
		return nil
	},
}

var apiKeyDeleteCmd = &cobra.Command{
	Use:   "delete <api-key-id> [api-key-id...]",
	Short: "Delete API keys",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.APIKeyMutationResponse
		input := map[string]interface{}{"ids": args}
		if err := queryWithTimeout(api.DeleteAPIKeyMutation, map[string]interface{}{"input": input}, &resp); err != nil {
			return fmt.Errorf("failed to delete API key: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(map[string]bool{"deleted": resp.APIKey.Delete})
		}
		if resp.APIKey.Delete {
			out.Success("API key deleted")
		} else {
			out.Warn("API key was not deleted")
		}
		return nil
	},
}

var apiKeyRolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "List available API key roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.APIKeyMetadataResponse
		if err := queryWithTimeout(api.APIKeyMetadataQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list API key roles: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp.APIKeyPossibleRoles)
		}
		for _, role := range resp.APIKeyPossibleRoles {
			out.Println(role)
		}
		return nil
	},
}

var apiKeyPermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "List available API key permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resp api.APIKeyMetadataResponse
		if err := queryWithTimeout(api.APIKeyMetadataQuery, nil, &resp); err != nil {
			return fmt.Errorf("failed to list API key permissions: %w", err)
		}
		if out.IsJSON() {
			return out.JSON(resp)
		}
		rows := make([][]string, 0, len(resp.APIKeyPossiblePermissions))
		for _, permission := range resp.APIKeyPossiblePermissions {
			rows = append(rows, []string{permission.Resource, strings.Join(permission.Actions, ",")})
		}
		out.Table([]string{"RESOURCE", "ACTIONS"}, rows)
		return nil
	},
}

var apiKeyAddRoleCmd = &cobra.Command{
	Use:   "add-role <api-key-id>",
	Short: "Add a role to an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiKeyAddRole == "" {
			return fmt.Errorf("--role is required")
		}
		return runAPIKeyRoleMutation(api.AddAPIKeyRoleMutation, args[0], apiKeyAddRole, "addRole", "Role added")
	},
}

var apiKeyRemoveRoleCmd = &cobra.Command{
	Use:   "remove-role <api-key-id>",
	Short: "Remove a role from an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiKeyRemoveRole == "" {
			return fmt.Errorf("--role is required")
		}
		return runAPIKeyRoleMutation(api.RemoveAPIKeyRoleMutation, args[0], apiKeyRemoveRole, "removeRole", "Role removed")
	},
}

func apiKeyInput(name, description string, roles, permissions []string) (map[string]interface{}, error) {
	input := map[string]interface{}{}
	if name != "" {
		input["name"] = name
	}
	if description != "" {
		input["description"] = description
	}
	if len(roles) > 0 {
		input["roles"] = normalizeEnums(roles)
	}
	if len(permissions) > 0 {
		parsed, err := parsePermissions(permissions)
		if err != nil {
			return nil, err
		}
		input["permissions"] = parsed
	}
	return input, nil
}

func parsePermissions(values []string) ([]map[string]interface{}, error) {
	permissions := make([]map[string]interface{}, 0, len(values))
	for _, value := range values {
		resource, actions, ok := strings.Cut(value, ":")
		if !ok || resource == "" || actions == "" {
			return nil, fmt.Errorf("permission %q must use RESOURCE:ACTION[,ACTION]", value)
		}
		permissions = append(permissions, map[string]interface{}{
			"resource": strings.ToUpper(resource),
			"actions":  normalizeEnums(strings.Split(actions, ",")),
		})
	}
	return permissions, nil
}

func normalizeEnums(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, strings.ToUpper(value))
	}
	return normalized
}

func runAPIKeyRoleMutation(query, id, role, fieldName, success string) error {
	var resp api.APIKeyMutationResponse
	input := map[string]interface{}{
		"apiKeyId": id,
		"role":     strings.ToUpper(role),
	}
	if err := queryWithTimeout(query, map[string]interface{}{"input": input}, &resp); err != nil {
		return fmt.Errorf("failed to %s API key role: %w", fieldName, err)
	}

	ok := resp.APIKey.AddRole
	if fieldName == "removeRole" {
		ok = resp.APIKey.RemoveRole
	}
	if out.IsJSON() {
		return out.JSON(map[string]bool{"updated": ok})
	}
	if ok {
		out.Success(success)
	} else {
		out.Warn("API key role was not updated")
	}
	return nil
}

func renderAPIKeys(keys []api.APIKey) {
	rows := make([][]string, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, []string{
			key.ID,
			key.Name,
			strings.Join(key.Roles, ","),
			key.CreatedAt,
			key.Description,
		})
	}
	out.Table([]string{"ID", "NAME", "ROLES", "CREATED", "DESCRIPTION"}, rows)
}

func init() {
	apiKeyCmd.AddCommand(apiKeyListCmd)
	apiKeyCmd.AddCommand(apiKeyCreateCmd)
	apiKeyCmd.AddCommand(apiKeyUpdateCmd)
	apiKeyCmd.AddCommand(apiKeyDeleteCmd)
	apiKeyCmd.AddCommand(apiKeyRolesCmd)
	apiKeyCmd.AddCommand(apiKeyPermissionsCmd)
	apiKeyCmd.AddCommand(apiKeyAddRoleCmd)
	apiKeyCmd.AddCommand(apiKeyRemoveRoleCmd)

	apiKeyCreateCmd.Flags().StringVar(&apiKeyCreateName, "name", "", "API key name")
	apiKeyCreateCmd.Flags().StringVar(&apiKeyCreateDescription, "description", "", "API key description")
	apiKeyCreateCmd.Flags().StringSliceVar(&apiKeyCreateRoles, "role", nil, "role to grant; repeat or comma-separate")
	apiKeyCreateCmd.Flags().StringArrayVar(&apiKeyCreatePermissions, "permission", nil, "permission in RESOURCE:ACTION[,ACTION] form; repeatable")
	apiKeyCreateCmd.Flags().BoolVar(&apiKeyCreateOverwrite, "overwrite", false, "replace an existing key with the same name")

	apiKeyUpdateCmd.Flags().StringVar(&apiKeyUpdateName, "name", "", "new API key name")
	apiKeyUpdateCmd.Flags().StringVar(&apiKeyUpdateDescription, "description", "", "new API key description")
	apiKeyUpdateCmd.Flags().StringSliceVar(&apiKeyUpdateRoles, "role", nil, "replacement role list; repeat or comma-separate")
	apiKeyUpdateCmd.Flags().StringArrayVar(&apiKeyUpdatePermissions, "permission", nil, "replacement permission in RESOURCE:ACTION[,ACTION] form; repeatable")

	apiKeyAddRoleCmd.Flags().StringVar(&apiKeyAddRole, "role", "", "role to add")
	apiKeyRemoveRoleCmd.Flags().StringVar(&apiKeyRemoveRole, "role", "", "role to remove")
}
