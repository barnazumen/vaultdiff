package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	namespacesCmd := &cobra.Command{
		Use:   "namespaces",
		Short: "List Vault namespaces",
		Long:  "List all child namespaces under a given prefix. Uses VAULT_ADDR and VAULT_TOKEN environment variables.",
		RunE:  runListNamespaces,
	}
	namespacesCmd.Flags().String("prefix", "", "Namespace prefix to list under (optional)")
	rootCmd.AddCommand(namespacesCmd)
}

func runListNamespaces(cmd *cobra.Command, _ []string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	prefix, _ := cmd.Flags().GetString("prefix")

	client := vault.NewClient(addr, token)
	namespaces, err := client.ListNamespaces(prefix)
	if err != nil {
		return fmt.Errorf("listing namespaces: %w", err)
	}

	if len(namespaces) == 0 {
		fmt.Println("No namespaces found.")
		return nil
	}

	sort.Slice(namespaces, func(i, j int) bool {
		return namespaces[i].Path < namespaces[j].Path
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tID\tMETADATA KEYS")
	for _, ns := range namespaces {
		metaCount := len(ns.CustomMetadata)
		fmt.Fprintf(w, "%s\t%s\t%d\n", ns.Path, ns.ID, metaCount)
	}
	return w.Flush()
}
