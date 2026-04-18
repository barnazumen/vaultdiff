package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultdiff/internal/diff"
	"github.com/your-org/vaultdiff/internal/vault"
)

var (
	diffVersionFrom int
	diffVersionTo   int
	diffMount       string
	diffNoColor     bool
	diffShowAll     bool
)

func init() {
	cmd := &cobra.Command{
		Use:   "diff <secret-path>",
		Short: "Diff two versions of a Vault secret",
		Args:  cobra.ExactArgs(1),
		RunE:  runDiffVersions,
	}
	cmd.Flags().IntVar(&diffVersionFrom, "from", 0, "Source version (required)")
	cmd.Flags().IntVar(&diffVersionTo, "to", 0, "Target version (required)")
	cmd.Flags().StringVar(&diffMount, "mount", "secret", "KV mount path")
	cmd.Flags().BoolVar(&diffNoColor, "no-color", false, "Disable color output")
	cmd.Flags().BoolVar(&diffShowAll, "show-all", false, "Show unchanged keys too")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	rootCmd.AddCommand(cmd)
}

func runDiffVersions(cmd *cobra.Command, args []string) error {
	secretPath := args[0]

	client, err := vault.NewClient(
		mustEnv("VAULT_ADDR"),
		mustEnv("VAULT_TOKEN"),
	)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	changes, err := vault.DiffVersions(client, diffMount, secretPath, diffVersionFrom, diffVersionTo)
	if err != nil {
		return fmt.Errorf("diffing versions: %w", err)
	}

	opts := diff.RenderOptions{
		NoColor:       diffNoColor,
		ShowUnchanged: diffShowAll,
	}
	diff.Render(os.Stdout, changes, opts)
	return nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "error: %s is not set\n", key)
		os.Exit(1)
	}
	return v
}
