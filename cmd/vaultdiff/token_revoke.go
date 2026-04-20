package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var selfRevoke bool
	var targetToken string

	cmd := &cobra.Command{
		Use:   "token-revoke",
		Short: "Revoke a Vault token",
		Long:  "Revoke a specific Vault token or self-revoke the currently authenticated token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTokenRevoke(selfRevoke, targetToken)
		},
	}

	cmd.Flags().BoolVar(&selfRevoke, "self", false, "Revoke the token used to authenticate (self-revoke)")
	cmd.Flags().StringVar(&targetToken, "token", "", "Token to revoke (required unless --self is set)")

	rootCmd.AddCommand(cmd)
}

func runTokenRevoke(selfRevoke bool, targetToken string) error {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		return fmt.Errorf("VAULT_ADDR environment variable is not set")
	}
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return fmt.Errorf("VAULT_TOKEN environment variable is not set")
	}
	if !selfRevoke && targetToken == "" {
		return fmt.Errorf("--token is required when --self is not set")
	}

	client := vault.NewClient(addr, token)
	if err := client.RevokeToken(targetToken, selfRevoke); err != nil {
		return fmt.Errorf("token revoke failed: %w", err)
	}

	if selfRevoke {
		fmt.Println("Self-revoke successful.")
	} else {
		fmt.Printf("Token %q has been revoked.\n", targetToken)
	}
	return nil
}
