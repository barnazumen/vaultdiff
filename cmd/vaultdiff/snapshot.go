package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/vaultdiff/internal/vault"
)

func init() {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "snapshot <path>",
		Short: "Export all versions of a secret to a JSON snapshot file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnapshot(args[0], outputFile)
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: <path>.snapshot.json)")
	rootCmd.AddCommand(cmd)
}

func runSnapshot(path, outputFile string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	if outputFile == "" {
		outputFile = path + ".snapshot.json"
	}

	client := vault.NewClient(addr, token)

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	if err := client.ExportSnapshot(path, f); err != nil {
		return fmt.Errorf("export snapshot: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Snapshot written to %s\n", outputFile)
	return nil
}
