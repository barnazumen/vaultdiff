package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultdiff/internal/audit"
	"vaultdiff/internal/diff"
	"vaultdiff/internal/vault"
)

func runDiff(_ *cobra.Command, _ []string) error {
	addr := vaultAddr
	if addr == "" {
		addr = os.Getenv("VAULT_ADDR")
	}
	token := vaultToken
	if token == "" {
		token = os.Getenv("VAULT_TOKEN")
	}

	client, err := vault.NewClient(addr, token)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	secretA, err := client.ReadSecretVersion(mount, path, versionA)
	if err != nil {
		return fmt.Errorf("reading version %d: %w", versionA, err)
	}

	secretB, err := client.ReadSecretVersion(mount, path, versionB)
	if err != nil {
		return fmt.Errorf("reading version %d: %w", versionB, err)
	}

	changes := diff.Compare(secretA, secretB)

	output := diff.Render(changes, diff.RenderOptions{
		NoColor:       noColor,
		ShowUnchanged: showAll,
	})
	fmt.Print(output)

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer f.Close()

		logger := audit.NewLogger(f)
		if err := logger.Log(path, versionA, versionB, changes); err != nil {
			return fmt.Errorf("writing audit log: %w", err)
		}
	}

	return nil
}
