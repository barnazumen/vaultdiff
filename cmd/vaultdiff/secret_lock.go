package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"vaultdiff/internal/audit"
	"vaultdiff/internal/vault"
)

func init() {
	var mount string
	var reason string
	var ttl int

	lockCmd := &cobra.Command{
		Use:   "lock <path>",
		Short: "Lock a secret path to prevent modifications",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLockSecret(args[0], mount, reason, ttl)
		},
	}
	lockCmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	lockCmd.Flags().StringVar(&reason, "reason", "", "Reason for locking")
	lockCmd.Flags().IntVar(&ttl, "ttl", 0, "Lock TTL in seconds (0 = no expiry)")

	unlockCmd := &cobra.Command{
		Use:   "unlock <path>",
		Short: "Unlock a previously locked secret path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlockSecret(args[0], mount)
		},
	}
	unlockCmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")

	rootCmd.AddCommand(lockCmd, unlockCmd)
}

func runLockSecret(path, mount, reason string, ttl int) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")
	c := vault.NewClient(addr)

	info, err := c.LockSecret(path, mount, token, reason, ttl)
	if err != nil {
		return fmt.Errorf("failed to lock secret: %w", err)
	}

	fmt.Printf("Locked: %s (by %s at %s)\n", info.Path, info.LockedBy, info.LockedAt.Format("2006-01-02T15:04:05Z"))
	if !info.ExpiresAt.IsZero() {
		fmt.Printf("Expires: %s\n", info.ExpiresAt.Format("2006-01-02T15:04:05Z"))
	}

	if logPath := os.Getenv("VAULTDIFF_AUDIT_LOG"); logPath != "" {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()
		_ = audit.LogLockEvent(f, "lock", path, mount, token, reason, ttl)
	}
	return nil
}

func runUnlockSecret(path, mount string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")
	c := vault.NewClient(addr)

	if err := c.UnlockSecret(path, mount, token); err != nil {
		return fmt.Errorf("failed to unlock secret: %w", err)
	}
	fmt.Printf("Unlocked: %s\n", path)

	if logPath := os.Getenv("VAULTDIFF_AUDIT_LOG"); logPath != "" {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()
		_ = audit.LogLockEvent(f, "unlock", path, mount, token, "", 0)
	}
	return nil
}

// ensure strconv import is used if needed for future TTL formatting
var _ = strconv.Itoa
