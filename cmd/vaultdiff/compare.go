package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"vaultdiff/internal/diff"
	"vaultdiff/internal/vault"
)

var compareCmd = &cobra.Command{
	Use:   "compare <mount> <path>",
	Short: "Compare two specific versions of a secret",
	Args:  cobra.ExactArgs(2),
	RunE:  runCompare,
}

var (
	cmpFrom int
	cmpTo   int
)

func init() {
	compareCmd.Flags().IntVar(&cmpFrom, "from", 0, "Source version (0 = auto)")
	compareCmd.Flags().IntVar(&cmpTo, "to", 0, "Target version (0 = latest)")
	rootCmd.AddCommand(compareCmd)
}

func runCompare(cmd *cobra.Command, args []string) error {
	mount := args[0]
	path := args[1]

	client, err := vault.NewClient(os.Getenv("VAULT_ADDR"), os.Getenv("VAULT_TOKEN"))
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	versions, err := client.ListVersions(mount, path)
	if err != nil {
		return fmt.Errorf("listing versions: %w", err)
	}

	pair, err := vault.ResolveVersionPair(versions, vault.VersionPair{From: cmpFrom, To: cmpTo})
	if err != nil {
		return fmt.Errorf("resolving versions: %w", err)
	}

	from, to, err := client.ReadVersionPair(mount, path, pair)
	if err != nil {
		return err
	}

	changes := diff.Compare(from, to)
	fmt.Printf("Comparing %s@v%s..v%s\n", path,
		strconv.Itoa(pair.From), strconv.Itoa(pair.To))
	diff.Render(changes, true, false)
	printSummary(diff.Summarize(changes))
	return nil
}
