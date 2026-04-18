package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var mount string

	cmd := &cobra.Command{
		Use:   "rollback <path> <version>",
		Short: "Roll back a secret to a previous version",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRollback(mount, args[0], args[1])
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(cmd)
}

func runRollback(mount, secretPath, versionStr string) error {
	version, err := strconv.Atoi(versionStr)
	if err != nil || version < 1 {
		return fmt.Errorf("invalid version %q: must be a positive integer", versionStr)
	}

	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")
	client := vault.NewClient(addr, token)

	result, err := client.RollbackToVersion(mount, secretPath, version)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Rolled back %s: version %d → %d (new version %d)\n",
		result.Path, version, result.FromVersion, result.ToVersion)
	return nil
}
