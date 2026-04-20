package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	rootCmd.AddCommand(tokenRenewCmd)
}

var tokenRenewCmd = &cobra.Command{
	Use:   "token-renew",
	Short: "Renew the current Vault token",
	Long:  "Renews the Vault token specified via VAULT_TOKEN and prints the updated lease duration.",
	RunE:  runTokenRenew,
}

func runTokenRenew(cmd *cobra.Command, args []string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	client := vault.NewClient(addr, token)
	result, err := client.RenewToken(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error renewing token: %v\n", err)
		return err
	}

	fmt.Printf("Token renewed successfully.\n")
	fmt.Printf("  Client Token   : %s\n", result.ClientToken)
	fmt.Printf("  Lease Duration : %d seconds\n", result.LeaseDuration)
	fmt.Printf("  Renewable      : %v\n", result.Renewable)
	return nil
}
