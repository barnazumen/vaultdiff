package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/vaultdiff/internal/vault"
)

func init() {
	var tokenA, tokenB, auditLog string

	cmd := &cobra.Command{
		Use:   "policy-audit <policy-name-a> <policy-name-b>",
		Short: "Compare two Vault policies and write an audit log entry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPolicyAudit(args[0], args[1], tokenA, tokenB, auditLog)
		},
	}

	cmd.Flags().StringVar(&tokenA, "token-a", "", "Vault token for policy A (default: VAULT_TOKEN)")
	cmd.Flags().StringVar(&tokenB, "token-b", "", "Vault token for policy B (default: VAULT_TOKEN)")
	cmd.Flags().StringVar(&auditLog, "audit-log", "vaultdiff-policy-audit.jsonl", "Path to audit log file")

	rootCmd.AddCommand(cmd)
}

func runPolicyAudit(policyA, policyB, tokenA, tokenB, auditLog string) error {
	addr := mustEnv("VAULT_ADDR")
	if tokenA == "" {
		tokenA = mustEnv("VAULT_TOKEN")
	}
	if tokenB == "" {
		tokenB = tokenA
	}

	clientA := vault.NewClient(addr, tokenA)
	clientB := vault.NewClient(addr, tokenB)

	diff, err := vault.ComparePolicies(clientA, clientB, policyA, policyB)
	if err != nil {
		return fmt.Errorf("compare policies: %w", err)
	}

	fmt.Println(vault.FormatPolicyDiff(diff))

	if err := vault.LogPolicyAudit(auditLog, policyA, policyB, diff); err != nil {
		fmt.Fprintf(os.Stderr, "warning: audit log write failed: %v\n", err)
	}

	return nil
}
