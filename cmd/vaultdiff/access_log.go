package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultdiff/internal/audit"
	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var mount string
	var auditLog string

	cmd := &cobra.Command{
		Use:   "access-log <secret-path>",
		Short: "Show version access history for a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAccessLog(args[0], mount, auditLog)
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	cmd.Flags().StringVar(&auditLog, "audit-log", "", "Path to write audit log (optional)")

	rootCmd.AddCommand(cmd)
}

func runAccessLog(secretPath, mount, auditLogPath string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	client := vault.NewClient(addr, token)

	result, err := client.ReadAccessLog(mount, secretPath)
	if err != nil {
		return fmt.Errorf("reading access log: %w", err)
	}

	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].Version < result.Entries[j].Version
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tOPERATION\tTIMESTAMP")
	for _, entry := range result.Entries {
		fmt.Fprintf(w, "%d\t%s\t%s\n", entry.Version, entry.Operation, entry.Timestamp.Format("2006-01-02 15:04:05"))
	}
	w.Flush()

	if auditLogPath != "" {
		f, err := os.OpenFile(auditLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer f.Close()
		if err := audit.LogAccessLogEvent(f, result); err != nil {
			return fmt.Errorf("writing audit log: %w", err)
		}
	}

	return nil
}
