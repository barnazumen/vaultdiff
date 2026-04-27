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
		Use:   "audit-trail <secret-path>",
		Short: "Display the audit trail for a secret's versions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuditTrail(args[0], mount, auditLog)
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	cmd.Flags().StringVar(&auditLog, "audit-log", "", "Path to write audit log (optional)")

	rootCmd.AddCommand(cmd)
}

func runAuditTrail(secretPath, mount, auditLog string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	client := vault.NewClient(addr, token)
	trail, err := client.ReadAuditTrail(mount, secretPath)
	if err != nil {
		return fmt.Errorf("reading audit trail: %w", err)
	}

	sort.Slice(trail.Entries, func(i, j int) bool {
		return trail.Entries[i].Version < trail.Entries[j].Version
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tACTION\tTIMESTAMP\tPATH")
	for _, e := range trail.Entries {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
			e.Version,
			e.Action,
			e.Timestamp.Format("2006-01-02T15:04:05Z"),
			e.Path,
		)
	}
	w.Flush()

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer f.Close()
		if err := audit.LogAuditTrailEvent(f, mount, secretPath, len(trail.Entries)); err != nil {
			return fmt.Errorf("writing audit log: %w", err)
		}
	}
	return nil
}
