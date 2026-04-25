package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/audit"
	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	pinCmd := &cobra.Command{
		Use:   "pin <path> <version>",
		Short: "Pin a specific version of a KV v2 secret",
		Args:  cobra.ExactArgs(2),
		RunE:  runPinSecret,
	}
	pinCmd.Flags().String("mount", "secret", "KV v2 mount path")
	pinCmd.Flags().String("audit-log", "", "Path to audit log file")
	pinCmd.Flags().String("actor", "", "Actor performing the pin (for audit)")

	unpinCmd := &cobra.Command{
		Use:   "unpin <path>",
		Short: "Remove the pin from a KV v2 secret",
		Args:  cobra.ExactArgs(1),
		RunE:  runUnpinSecret,
	}
	unpinCmd.Flags().String("mount", "secret", "KV v2 mount path")
	unpinCmd.Flags().String("audit-log", "", "Path to audit log file")
	unpinCmd.Flags().String("actor", "", "Actor performing the unpin (for audit)")

	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)
}

func runPinSecret(cmd *cobra.Command, args []string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")
	mount, _ := cmd.Flags().GetString("mount")
	actor, _ := cmd.Flags().GetString("actor")
	auditLog, _ := cmd.Flags().GetString("audit-log")

	secretPath := args[0]
	version, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid version %q: %w", args[1], err)
	}

	result, err := vault.PinSecret(addr, token, mount, secretPath, version)
	if err != nil {
		return err
	}

	fmt.Printf("Pinned %s at version %d\n", result.Path, result.Version)

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()
		_ = audit.LogPinEvent(f, mount, secretPath, version, true, actor)
	}
	return nil
}

func runUnpinSecret(cmd *cobra.Command, args []string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")
	mount, _ := cmd.Flags().GetString("mount")
	actor, _ := cmd.Flags().GetString("actor")
	auditLog, _ := cmd.Flags().GetString("audit-log")

	secretPath := args[0]

	result, err := vault.UnpinSecret(addr, token, mount, secretPath)
	if err != nil {
		return err
	}

	fmt.Printf("Unpinned %s\n", result.Path)

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()
		_ = audit.LogPinEvent(f, mount, secretPath, 0, false, actor)
	}
	return nil
}
