package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	vaultAddr  string
	vaultToken string
	mount      string
	path       string
	versionA   int
	versionB   int
	noColor    bool
	showAll    bool
	auditLog   string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "vaultdiff",
	Short: "Diff and audit changes between HashiCorp Vault secret versions",
	RunE:  runDiff,
}

func init() {
	rootCmd.Flags().StringVar(&vaultAddr, "addr", "", "Vault server address (overrides VAULT_ADDR)")
	rootCmd.Flags().StringVar(&vaultToken, "token", "", "Vault token (overrides VAULT_TOKEN)")
	rootCmd.Flags().StringVar(&mount, "mount", "secret", "KV v2 mount path")
	rootCmd.Flags().StringVar(&path, "path", "", "Secret path (required)")
	rootCmd.Flags().IntVar(&versionA, "version-a", 0, "First version to compare (required)")
	rootCmd.Flags().IntVar(&versionB, "version-b", 0, "Second version to compare (required)")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.Flags().BoolVar(&showAll, "show-all", false, "Show unchanged keys as well")
	rootCmd.Flags().StringVar(&auditLog, "audit-log", "", "Append audit entry to this file")
	_ = rootCmd.MarkFlagRequired("path")
	_ = rootCmd.MarkFlagRequired("version-a")
	_ = rootCmd.MarkFlagRequired("version-b")
}
