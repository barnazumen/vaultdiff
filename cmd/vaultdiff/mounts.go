package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	mountsCmd := &cobra.Command{
		Use:   "mounts",
		Short: "List all secret engine mounts in Vault",
		RunE:  runMounts,
	}
	rootCmd.AddCommand(mountsCmd)
}

func runMounts(cmd *cobra.Command, args []string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	c := vault.NewClient(addr, token)
	mounts, err := c.ListMounts()
	if err != nil {
		return fmt.Errorf("failed to list mounts: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tTYPE\tDESCRIPTION\tACCESSOR")
	for path, info := range mounts {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", path, info.Type, info.Description, info.Accessor)
	}
	return w.Flush()
}
