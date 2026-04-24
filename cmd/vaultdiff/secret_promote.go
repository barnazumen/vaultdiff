package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"vaultdiff/internal/audit"
	"vaultdiff/internal/vault"
)

var (
	promoteSrcMount string
	promoteDstMount string
	promotePath     string
	promoteDstPath  string
	promoteAuditLog string
)

func init() {
	promoteCmd := &cobra.Command{
		Use:   "promote",
		Short: "Promote a secret from one mount to another",
		RunE:  runPromoteSecret,
	}

	promoteCmd.Flags().StringVar(&promoteSrcMount, "src-mount", "", "Source KV mount (required)")
	promoteCmd.Flags().StringVar(&promoteDstMount, "dst-mount", "", "Destination KV mount (required)")
	promoteCmd.Flags().StringVar(&promotePath, "path", "", "Secret path in source mount (required)")
	promoteCmd.Flags().StringVar(&promoteDstPath, "dst-path", "", "Secret path in destination mount (defaults to --path)")
	promoteCmd.Flags().StringVar(&promoteAuditLog, "audit-log", "", "Path to append audit log entry")

	_ = promoteCmd.MarkFlagRequired("src-mount")
	_ = promoteCmd.MarkFlagRequired("dst-mount")
	_ = promoteCmd.MarkFlagRequired("path")

	rootCmd.AddCommand(promoteCmd)
}

func runPromoteSecret(cmd *cobra.Command, args []string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	result, err := vault.PromoteSecret(addr, token, promoteSrcMount, promoteDstMount, promotePath, promoteDstPath)
	if err != nil {
		return fmt.Errorf("promote failed: %w", err)
	}

	sort.Strings(result.Keys)
	fmt.Printf("Promoted secret '%s' from mount '%s' (v%d) → mount '%s'\n",
		result.Path, result.SourceMount, result.Version, result.DestMount)
	fmt.Printf("Keys promoted (%d): %v\n", len(result.Keys), result.Keys)

	if promoteAuditLog != "" {
		f, err := os.OpenFile(promoteAuditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()
		if err := audit.LogPromoteEvent(f, result.SourceMount, result.DestMount, result.Path, result.Version, result.Keys); err != nil {
			return fmt.Errorf("write audit log: %w", err)
		}
	}

	return nil
}
