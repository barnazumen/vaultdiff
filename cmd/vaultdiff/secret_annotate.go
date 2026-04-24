package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultdiff/internal/audit"
	"github.com/yourusername/vaultdiff/internal/vault"
)

func init() {
	var mount, auditLog string

	setCmd := &cobra.Command{
		Use:   "annotate [path] [key] [value]",
		Short: "Set an annotation on a secret",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetAnnotation(args[0], args[1], args[2], mount, auditLog)
		},
	}
	setCmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")
	setCmd.Flags().StringVar(&auditLog, "audit-log", "", "path to audit log file")

	getCmd := &cobra.Command{
		Use:   "annotations [path]",
		Short: "Get annotations on a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetAnnotations(args[0], mount)
		},
	}
	getCmd.Flags().StringVar(&mount, "mount", "secret", "KV mount path")

	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
}

func runSetAnnotation(path, key, value, mount, auditLog string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	client := vault.NewClient(addr, token)
	if err := client.SetAnnotation(mount, path, key, value); err != nil {
		return fmt.Errorf("set annotation: %w", err)
	}

	fmt.Printf("Annotation set: %s=%s on %s/%s\n", key, value, mount, path)

	if auditLog != "" {
		f, err := os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()
		if err := audit.LogAnnotateEvent(f, mount, path, key, value); err != nil {
			return fmt.Errorf("write audit log: %w", err)
		}
	}
	return nil
}

func runGetAnnotations(path, mount string) error {
	addr := mustEnv("VAULT_ADDR")
	token := mustEnv("VAULT_TOKEN")

	client := vault.NewClient(addr, token)
	result, err := client.GetAnnotations(mount, path)
	if err != nil {
		return fmt.Errorf("get annotations: %w", err)
	}

	if len(result.Annotations) == 0 {
		fmt.Println("No annotations found.")
		return nil
	}

	fmt.Printf("Annotations for %s/%s:\n", mount, path)
	for k, v := range result.Annotations {
		fmt.Printf("  %s = %s\n", k, v)
	}
	return nil
}
