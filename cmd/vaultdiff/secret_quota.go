package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/audit"
	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var quotaName string
	var auditLog string

	cmd := &cobra.Command{
		Use:   "quota",
		Short: "Read quota information for a named Vault rate-limit quota",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSecretQuota(quotaName, auditLog)
		},
	}

	cmd.Flags().StringVar(&quotaName, "name", "", "Name of the quota to inspect (required)")
	cmd.Flags().StringVar(&auditLog, "audit-log", "", "Path to append audit log entry (optional)")
	_ = cmd.MarkFlagRequired("name")

	rootCmd.AddCommand(cmd)
}

func runSecretQuota(quotaName, auditLog string) error {
	vaultAddr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")

	if vaultAddr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	info, err := vault.ReadSecretQuota(vaultAddr, token, quotaName)
	if err != nil {
		return fmt.Errorf("reading quota: %w", err)
	}

	fmt.Printf("Quota:           %s\n", quotaName)
	fmt.Printf("Path:            %s\n", info.Path)
	fmt.Printf("Type:            %s\n", info.Type)
	fmt.Printf("Max Leases:      %d\n", info.MaxLeases)
	fmt.Printf("Current Leases:  %d\n", info.CurrentLeases)
	fmt.Printf("Rate:            %.2f req/s\n", info.Rate)
	fmt.Printf("Burst:           %d\n", info.Burst)
	fmt.Printf("Interval:        %ds\n", info.IntervalSeconds)

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer f.Close()

		if err := audit.LogQuotaEvent(f, quotaName, info.Path, info.Type, info.MaxLeases, info.CurrentLeases, info.Rate, info.Burst); err != nil {
			return fmt.Errorf("writing audit log: %w", err)
		}
	}

	return nil
}
