package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"vaultdiff/internal/audit"
	"vaultdiff/internal/vault"
)

func init() {
	var requiredKeys string
	var version int
	var strict bool
	var auditLog string

	cmd := &cobra.Command{
		Use:   "validate <mount> <secret-path>",
		Short: "Validate that a secret version contains required keys",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidateSecret(args[0], args[1], version, requiredKeys, strict, auditLog)
		},
	}
	cmd.Flags().StringVar(&requiredKeys, "keys", "", "Comma-separated list of required keys")
	cmd.Flags().IntVar(&version, "version", 1, "Secret version to validate")
	cmd.Flags().BoolVar(&strict, "strict", false, "Fail if unexpected keys are present")
	cmd.Flags().StringVar(&auditLog, "audit-log", "", "Path to append audit log (optional)")
	_ = cmd.MarkFlagRequired("keys")
	rootCmd.AddCommand(cmd)
}

func runValidateSecret(mount, secretPath string, version int, keysFlag string, strict bool, auditLog string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	requiredKeys := strings.Split(keysFlag, ",")
	for i, k := range requiredKeys {
		requiredKeys[i] = strings.TrimSpace(k)
	}

	res, err := vault.ValidateSecretKeys(addr, token, mount, secretPath, version, requiredKeys, strict)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer f.Close()
		if err := audit.LogValidateEvent(f, res.Path, res.Version, res.Valid, res.Missing, res.Extra, strict); err != nil {
			return fmt.Errorf("writing audit log: %w", err)
		}
	}

	fmt.Printf("Path:    %s\n", res.Path)
	fmt.Printf("Version: %s\n", strconv.Itoa(res.Version))
	if res.Valid {
		fmt.Println("Status:  ✓ VALID")
	} else {
		fmt.Println("Status:  ✗ INVALID")
		if len(res.Missing) > 0 {
			fmt.Printf("Missing: %s\n", strings.Join(res.Missing, ", "))
		}
		if len(res.Extra) > 0 {
			fmt.Printf("Extra:   %s\n", strings.Join(res.Extra, ", "))
		}
		return fmt.Errorf("secret %s@v%d failed validation", res.Path, res.Version)
	}
	return nil
}
