package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var snapshotFile string

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore secret versions from a snapshot file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestore(snapshotFile)
		},
	}

	cmd.Flags().StringVarP(&snapshotFile, "file", "f", "", "Path to snapshot file (required)")
	_ = cmd.MarkFlagRequired("file")

	rootCmd.AddCommand(cmd)
}

func runRestore(snapshotFile string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")
	mount := os.Getenv("VAULT_MOUNT")
	if mount == "" {
		mount = "secret"
	}
	path := mustEnv("VAULT_PATH")

	snapshot, err := vault.LoadSnapshot(snapshotFile)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	client := vault.NewClient(addr, token)

	if err := client.RestoreFromSnapshot(mount, path, snapshot); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Printf("Restored %d version(s) to %s/%s\n", len(snapshot), mount, path)
	return nil
}
