package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/diff"
)

var showSummary bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&showSummary, "summary", false, "print a change summary after the diff")
}

// printSummary writes a diff Summary to stdout if the --summary flag is set.
func printSummary(_ *cobra.Command, changes []diff.ChangeRecord) {
	if !showSummary {
		return
	}
	s := diff.Summarize(changes)
	fmt.Fprintln(os.Stdout, s.String())
}
