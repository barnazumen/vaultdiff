package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/vault"
)

func init() {
	var addr1, addr2, token1, token2 string

	cmd := &cobra.Command{
		Use:   "mounts-diff",
		Short: "Diff secret engine mounts between two Vault instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMountsDiff(addr1, addr2, token1, token2)
		},
	}

	cmd.Flags().StringVar(&addr1, "addr1", "", "First Vault address (required)")
	cmd.Flags().StringVar(&addr2, "addr2", "", "Second Vault address (required)")
	cmd.Flags().StringVar(&token1, "token1", "", "Token for first Vault (required)")
	cmd.Flags().StringVar(&token2, "token2", "", "Token for second Vault (required)")

	_ = cmd.MarkFlagRequired("addr1")
	_ = cmd.MarkFlagRequired("addr2")
	_ = cmd.MarkFlagRequired("token1")
	_ = cmd.MarkFlagRequired("token2")

	rootCmd.AddCommand(cmd)
}

func runMountsDiff(addr1, addr2, token1, token2 string) error {
	c1, err := vault.NewClient(addr1, token1)
	if err != nil {
		return fmt.Errorf("client1: %w", err)
	}
	c2, err := vault.NewClient(addr2, token2)
	if err != nil {
		return fmt.Errorf("client2: %w", err)
	}

	mounts1, err := vault.ListMounts(c1)
	if err != nil {
		return fmt.Errorf("list mounts (addr1): %w", err)
	}
	mounts2, err := vault.ListMounts(c2)
	if err != nil {
		return fmt.Errorf("list mounts (addr2): %w", err)
	}

	diffs := vault.DiffMounts(mounts1, mounts2)
	if len(diffs) == 0 {
		fmt.Println("No differences found.")
		return nil
	}

	fmt.Fprintln(os.Stdout, vault.FormatMountDiff(diffs))
	return nil
}
