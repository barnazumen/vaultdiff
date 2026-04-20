package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var path string

	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "Show token capabilities on a Vault secret path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTokenCapabilities(path)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Vault secret path to check capabilities for (required)")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}

func runTokenCapabilities(path string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	client := vault.NewClient(addr, token)
	caps, err := client.GetTokenCapabilities(path)
	if err != nil {
		return fmt.Errorf("failed to get capabilities: %w", err)
	}

	fmt.Printf("Path:         %s\n", caps.Path)
	fmt.Printf("Capabilities: %s\n", strings.Join(caps.Capabilities, ", "))
	return nil
}
