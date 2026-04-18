package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	var mount string

	cmd := &cobra.Command{
		Use:   "metadata <path>",
		Short: "Show version metadata for a KV v2 secret path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetadata(args[0], mount)
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV v2 mount path")
	rootCmd.AddCommand(cmd)
}

func runMetadata(secretPath, mount string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	client := vault.NewClient(addr, token)
	meta, err := client.ReadSecretMetadata(mount, secretPath)
	if err != nil {
		return fmt.Errorf("reading metadata: %w", err)
	}

	fmt.Printf("Path:            %s\n", meta.Path)
	fmt.Printf("Current Version: %d\n", meta.CurrentVersion)
	fmt.Printf("Oldest Version:  %d\n", meta.OldestVersion)
	fmt.Printf("Created:         %s\n\n", meta.CreatedTime.Format("2006-01-02 15:04:05 UTC"))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VERSION\tCREATED\tDESTROYED")
	for key, v := range meta.Versions {
		created := v.CreatedTime.Format("2006-01-02 15:04:05")
		if v.CreatedTime.IsZero() {
			created = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%v\n", key, created, v.Destroyed)
	}
	return w.Flush()
}
