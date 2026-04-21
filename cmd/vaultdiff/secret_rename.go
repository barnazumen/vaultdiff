package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	var mount string

	cmd := &cobra.Command{
		Use:   "rename <src-path> <dst-path>",
		Short: "Rename a KV v2 secret by copying it to a new path and deleting the original",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRenameSecret(args[0], args[1], mount)
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(cmd)
}

func runRenameSecret(srcPath, dstPath, mount string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	client := newVaultClient(addr, token)
	if err := client.RenameSecret(rootCmd.Context(), mount, srcPath, dstPath); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}

	fmt.Printf("Renamed %q → %q on mount %q\n", srcPath, dstPath, mount)
	return nil
}
