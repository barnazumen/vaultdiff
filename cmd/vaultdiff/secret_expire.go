package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultdiff/internal/audit"
	"github.com/yourusername/vaultdiff/internal/vault"
)

var expireCmd = &cobra.Command{
	Use:   "expire",
	Short: "Check if a secret version is older than a given number of days",
	RunE:  runExpireCheck,
}

func init() {
	expireCmd.Flags().String("mount", "secret", "KV mount path")
	expireCmd.Flags().String("path", "", "Secret path (required)")
	expireCmd.Flags().Int("max-age", 30, "Maximum allowed age in days")
	expireCmd.Flags().String("audit-log", "", "Optional path to write audit log")
	_ = expireCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(expireCmd)
}

func runExpireCheck(cmd *cobra.Command, args []string) error {
	mount, _ := cmd.Flags().GetString("mount")
	path, _ := cmd.Flags().GetString("path")
	maxAge, _ := cmd.Flags().GetInt("max-age")
	auditLog, _ := cmd.Flags().GetString("audit-log")

	vaultAddr := mustEnv("VAULT_ADDR")
	vaultToken := mustEnv("VAULT_TOKEN")

	client := vault.NewClient(vaultAddr, vaultToken)
	expiry, err := client.CheckSecretExpiry(mount, path, maxAge)
	if err != nil {
		return fmt.Errorf("checking expiry: %w", err)
	}

	fmt.Printf("Path:        %s\n", expiry.Path)
	fmt.Printf("Version:     %d\n", expiry.Version)
	fmt.Printf("Created:     %s\n", expiry.CreatedTime.Format("2006-01-02"))
	fmt.Printf("Days Old:    %s\n", strconv.Itoa(expiry.DaysOld))
	fmt.Printf("Max Age:     %d days\n", maxAge)
	if expiry.Expired {
		fmt.Println("Status:      EXPIRED")
	} else {
		fmt.Println("Status:      OK")
	}

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer f.Close()
		if err := audit.LogExpireEvent(f, expiry, maxAge); err != nil {
			return fmt.Errorf("writing audit log: %w", err)
		}
	}

	if expiry.Expired {
		os.Exit(1)
	}
	return nil
}
