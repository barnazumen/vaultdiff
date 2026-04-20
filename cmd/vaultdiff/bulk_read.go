package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	var mount string
	var version int
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "bulk-read [path1] [path2] ...",
		Short: "Read multiple secrets concurrently",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBulkRead(args, mount, version, outputJSON)
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	cmd.Flags().IntVar(&version, "version", 0, "Secret version (0 = latest)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output results as JSON")

	rootCmd.AddCommand(cmd)
}

func runBulkRead(paths []string, mount string, version int, outputJSON bool) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	client := vault.NewClient(addr, token)
	results := client.ReadSecretsBulk(mount, paths, version)

	var errs []string

	if outputJSON {
		out := make(map[string]interface{})
		for _, r := range results {
			if r.Error != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", r.Path, r.Error))
				continue
			}
			out[r.Path] = r.Data
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			return fmt.Errorf("encode JSON: %w", err)
		}
	} else {
		for _, r := range results {
			if r.Error != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", r.Path, r.Error))
				continue
			}
			fmt.Printf("=== %s ===\n", r.Path)
			for k, v := range r.Data {
				fmt.Printf("  %s = %v\n", k, v)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors reading secrets:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}
