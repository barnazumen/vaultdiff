package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

var secretMoveCmd = &cobra.Command{
	Use:   "move <mount> <source> <destination>",
	Short: "Move a KV secret from one path to another (copy + delete source)",
	Args:  cobra.ExactArgs(3),
	RunE:  runMoveSecret,
}

func init() {
	rootCmd.AddCommand(secretMoveCmd)
}

func runMoveSecret(cmd *cobra.Command, args []string) error {
	mount := args[0]
	src := args[1]
	dst := args[2]

	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		return fmt.Errorf("VAULT_ADDR environment variable is required")
	}

	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return fmt.Errorf("VAULT_TOKEN environment variable is required")
	}

	result, err := vault.MoveSecret(cmd.Context(), addr, token, mount, src, dst)
	if err != nil {
		return fmt.Errorf("move secret: %w", err)
	}

	fmt.Printf("Moved secret:\n")
	fmt.Printf("  Source:      %s/%s\n", mount, result.Source)
	fmt.Printf("  Destination: %s/%s\n", mount, result.Destination)
	fmt.Printf("  Versions copied: %d\n", result.Versions)
	return nil
}
