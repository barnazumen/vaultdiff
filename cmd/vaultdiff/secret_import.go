package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	vaultaudit "github.com/your-org/vaultdiff/internal/audit"
	"github.com/your-org/vaultdiff/internal/vault"
)

func init() {
	var mount string
	var auditLog string

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import secrets from a JSON file into Vault KV v2",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImportSecret(args[0], mount, auditLog)
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV v2 mount path")
	cmd.Flags().StringVar(&auditLog, "audit-log", "", "Path to write audit log (optional)")
	rootCmd.AddCommand(cmd)
}

func runImportSecret(file, mount, auditLog string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	results, err := vault.ImportSecretsFromFile(addr, token, mount, file)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	var succeeded, failed int
	for _, r := range results {
		if r.Success {
			succeeded++
			fmt.Printf("  ✓ %s\n", r.Path)
		} else {
			failed++
			fmt.Printf("  ✗ %s: %s\n", r.Path, r.Error)
		}
	}
	fmt.Printf("\nImport complete: %d succeeded, %d failed\n", succeeded, failed)

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()

		var auditResults []vaultaudit.ImportResult
		for _, r := range results {
			auditResults = append(auditResults, vaultaudit.ImportResult{
				Path:    r.Path,
				Success: r.Success,
				Error:   r.Error,
			})
		}
		if err := vaultaudit.LogImportEvent(f, mount, file, auditResults); err != nil {
			return fmt.Errorf("write audit log: %w", err)
		}
	}
	return nil
}
