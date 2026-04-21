package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	var srcMount, dstMount, srcPath, dstPath string
	var version int

	cmd := &cobra.Command{
		Use:   "kv-copy",
		Short: "Copy a KV-v2 secret from one path to another",
		Long: `Reads a secret from the source path (optionally at a pinned version)
 and writes its data to the destination path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKVCopy(srcMount, dstMount, srcPath, dstPath, version)
		},
	}

	cmd.Flags().StringVar(&srcMount, "src-mount", "secret", "Source KV mount")
	cmd.Flags().StringVar(&dstMount, "dst-mount", "", "Destination KV mount (defaults to src-mount)")
	cmd.Flags().StringVar(&srcPath, "src", "", "Source secret path (required)")
	cmd.Flags().StringVar(&dstPath, "dst", "", "Destination secret path (required)")
	cmd.Flags().IntVar(&version, "version", 0, "Source version to copy (0 = latest)")

	_ = cmd.MarkFlagRequired("src")
	_ = cmd.MarkFlagRequired("dst")

	rootCmd.AddCommand(cmd)
}

func runKVCopy(srcMount, dstMount, srcPath, dstPath string, version int) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" || token == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	client := vault.NewClient(addr, token)
	err := client.CopySecret(vault.KVCopyOptions{
		SourceMount: srcMount,
		DestMount:   dstMount,
		SourcePath:  srcPath,
		DestPath:    dstPath,
		Version:     version,
	})
	if err != nil {
		return fmt.Errorf("kv-copy failed: %w", err)
	}

	fmt.Printf("✔ Copied %s/%s → %s/%s\n",
		srcMount, srcPath,
		func() string {
			if dstMount == "" {
				return srcMount
			}
			return dstMount
		}(),
		dstPath,
	)
	return nil
}
