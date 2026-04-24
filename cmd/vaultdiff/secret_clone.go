package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"vaultdiff/internal/audit"
	"vaultdiff/internal/vault"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <source-path> <dest-path>",
	Short: "Clone a secret from one path to another",
	Args:  cobra.ExactArgs(2),
	RunE:  runCloneSecret,
}

func init() {
	cloneCmd.Flags().String("mount", "secret", "KV mount name")
	cloneCmd.Flags().Int("version", 0, "Source version to clone (0 = latest)")
	cloneCmd.Flags().Bool("overwrite", false, "Overwrite destination if it exists (PUT instead of POST)")
	cloneCmd.Flags().String("audit-log", "", "Path to write audit log entry")
	rootCmd.AddCommand(cloneCmd)
}

func runCloneSecret(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	dstPath := args[1]

	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	mount, _ := cmd.Flags().GetString("mount")
	versionStr, _ := cmd.Flags().GetInt("version")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	auditLog, _ := cmd.Flags().GetString("audit-log")

	result, err := vault.CloneSecret(addr, token, vault.CloneSecretOptions{
		SourcePath: srcPath,
		DestPath:   dstPath,
		Mount:      mount,
		Version:    versionStr,
		Overwrite:  overwrite,
	})
	if err != nil {
		return fmt.Errorf("clone failed: %w", err)
	}

	sort.Strings(result.Keys)
	fmt.Printf("Cloned %s (v%s) → %s\n", srcPath, strconv.Itoa(result.Version), dstPath)
	fmt.Printf("Keys cloned (%d): %v\n", len(result.Keys), result.Keys)

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()
		if err := audit.LogCloneEvent(f, srcPath, dstPath, mount, result.Version, result.Keys); err != nil {
			return fmt.Errorf("write audit log: %w", err)
		}
	}

	return nil
}
