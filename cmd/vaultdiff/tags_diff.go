package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/audit"
	"vaultdiff/internal/vault"
)

func init() {
	var auditLog string

	cmd := &cobra.Command{
		Use:   "tags-diff",
		Short: "Diff tags between two secret paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTagsDiff(auditLog)
		},
	}

	cmd.Flags().StringVar(&auditLog, "audit-log", "", "Path to audit log file")
	rootCmd.AddCommand(cmd)
}

func runTagsDiff(auditLogPath string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")
	srcPath := mustEnv("VAULT_SRC_PATH")
	dstPath := mustEnv("VAULT_DST_PATH")

	client := vault.NewClient(addr, token)

	srcTags, err := client.ReadSecretTags(srcPath)
	if err != nil {
		return fmt.Errorf("reading src tags: %w", err)
	}
	dstTags, err := client.ReadSecretTags(dstPath)
	if err != nil {
		return fmt.Errorf("reading dst tags: %w", err)
	}

	diffs := vault.DiffTags(srcTags, dstTags)
	fmt.Print(vault.FormatTagDiff(diffs))

	if auditLogPath != "" {
		added := map[string]string{}
		removed := map[string]string{}
		modified := map[string]string{}
		for _, d := range diffs {
			switch d.Status {
			case "added":
				added[d.Key] = d.NewVal
			case "removed":
				removed[d.Key] = d.OldVal
			case "modified":
				modified[d.Key] = d.NewVal
			}
		}
		f, err := os.OpenFile(auditLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer f.Close()
		if err := audit.LogTagEvent(f, srcPath, added, removed, modified); err != nil {
			return fmt.Errorf("writing audit log: %w", err)
		}
	}
	return nil
}
