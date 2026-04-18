package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/vaultdiff/internal/vault"
)

func init() {
	var path string

	cmd := &cobra.Command{
		Use:   "lease",
		Short: "Show lease info for a dynamic secret path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLeaseInfo(path)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Vault secret path (required)")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}

func runLeaseInfo(path string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	client := vault.NewClient(addr, token)
	info, err := client.ReadLeaseInfo(path)
	if err != nil {
		return fmt.Errorf("reading lease info: %w", err)
	}

	fmt.Printf("Lease ID:       %s\n", info.LeaseID)
	fmt.Printf("Renewable:      %v\n", info.Renewable)
	fmt.Printf("Lease Duration: %s\n", info.LeaseDuration)
	return nil
}
