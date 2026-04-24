package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Soft-delete (archive) a specific version of a KV v2 secret",
	RunE:  runArchiveSecret,
}

func init() {
	archiveCmd.Flags().String("path", "", "Secret path (required)")
	archiveCmd.Flags().String("mount", "secret", "KV v2 mount path")
	archiveCmd.Flags().Int("version", 0, "Version number to archive (required)")
	_ = archiveCmd.MarkFlagRequired("path")
	_ = archiveCmd.MarkFlagRequired("version")
	rootCmd.AddCommand(archiveCmd)
}

func runArchiveSecret(cmd *cobra.Command, _ []string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	path, _ := cmd.Flags().GetString("path")
	mount, _ := cmd.Flags().GetString("mount")
	version, _ := cmd.Flags().GetInt("version")

	if version <= 0 {
		return fmt.Errorf("--version must be a positive integer")
	}

	client := vault.NewClient(addr, token)
	result, err := client.ArchiveSecretVersion(mount, path, version)
	if err != nil {
		return fmt.Errorf("archive failed: %w", err)
	}

	fmt.Printf("Archived secret '%s' version %s\n",
		result.Path, strconv.Itoa(result.Version))
	return nil
}
