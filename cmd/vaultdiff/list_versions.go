package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	listCmd := &cobra.Command{
		Use:   "versions",
		Short: "List available versions of a secret",
		RunE:  runListVersions,
	}
	listCmd.Flags().String("mount", "secret", "KV mount path")
	listCmd.Flags().String("path", "", "Secret path (required)")
	listCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(listCmd)
}

func runListVersions(cmd *cobra.Command, _ []string) error {
	addr, _ := cmd.Root().PersistentFlags().GetString("address")
	token, _ := cmd.Root().PersistentFlags().GetString("token")
	mount, _ := cmd.Flags().GetString("mount")
	path, _ := cmd.Flags().GetString("path")

	client, err := vault.NewClient(addr, token)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	versions, err := client.ListVersions(mount, path)
	if err != nil {
		return fmt.Errorf("listing versions: %w", err)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version < versions[j].Version
	})

	fmt.Printf("Versions for %s/%s:\n\n", mount, path)
	fmt.Printf("  %-8s %-30s %-10s\n", "VERSION", "CREATED", "STATUS")
	for _, v := range versions {
		status := "active"
		if v.Destroyed {
			status = "destroyed"
		} else if v.DeletionTime != "" {
			status = "deleted"
		}
		fmt.Printf("  %-8d %-30s %-10s\n", v.Version, v.CreatedTime, status)
	}
	return nil
}
