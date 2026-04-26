package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	getCmd := &cobra.Command{
		Use:   "chmod-get <mount> <path>",
		Short: "Get access permissions for a secret",
		Args:  cobra.ExactArgs(2),
		RunE:  runGetChmod,
	}

	setCmd := &cobra.Command{
		Use:   "chmod-set <mount> <path>",
		Short: "Set access permissions for a secret",
		Args:  cobra.ExactArgs(2),
		RunE:  runSetChmod,
	}
	setCmd.Flags().String("owner", "", "Owner of the secret")
	setCmd.Flags().StringSlice("read-roles", nil, "Comma-separated roles with read access")
	setCmd.Flags().StringSlice("write-roles", nil, "Comma-separated roles with write access")
	setCmd.Flags().StringSlice("deny-roles", nil, "Comma-separated roles explicitly denied")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
}

func runGetChmod(cmd *cobra.Command, args []string) error {
	mount, path := args[0], args[1]
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	c := vault.NewClient(addr, token)
	perms, err := c.GetSecretPermissions(mount, path)
	if err != nil {
		return fmt.Errorf("get permissions: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Owner:       %s\n", perms.Owner)
	fmt.Fprintf(os.Stdout, "Read Roles:  %s\n", strings.Join(perms.ReadRoles, ", "))
	fmt.Fprintf(os.Stdout, "Write Roles: %s\n", strings.Join(perms.WriteRoles, ", "))
	fmt.Fprintf(os.Stdout, "Deny Roles:  %s\n", strings.Join(perms.DenyRoles, ", "))
	return nil
}

func runSetChmod(cmd *cobra.Command, args []string) error {
	mount, path := args[0], args[1]
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	owner, _ := cmd.Flags().GetString("owner")
	readRoles, _ := cmd.Flags().GetStringSlice("read-roles")
	writeRoles, _ := cmd.Flags().GetStringSlice("write-roles")
	denyRoles, _ := cmd.Flags().GetStringSlice("deny-roles")

	perms := vault.SecretPermissions{
		Owner:      owner,
		ReadRoles:  readRoles,
		WriteRoles: writeRoles,
		DenyRoles:  denyRoles,
	}

	c := vault.NewClient(addr, token)
	if err := c.SetSecretPermissions(mount, path, perms); err != nil {
		return fmt.Errorf("set permissions: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Permissions updated for %s/%s\n", mount, path)
	return nil
}
