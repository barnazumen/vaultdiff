package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

aultdiff/internal/audit/internal/vault() {
	w &cobra.Command{
	se:   "watch",
		Short: "Watch a secret path for version changes",
		RunE:  runWatch,
	}
	watchCmd.Flags().String("path", "", "Secret path to watch (required)")
	watchCmd.Flags().Duration("interval", 30*time.Second, "Poll interval")
	watchCmd.Flags().String("audit-log", "", "Path to audit log file")
	watchCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, _ []string) error {
	path, _ := cmd.Flags().GetString("path")
	interval, _ := cmd.Flags().GetDuration("interval")
	auditLog, _ := cmd.Flags().GetString("audit-log")

	c, err := vault.NewClient(mustEnv("VAULT_ADDR"), mustEnv("VAULT_TOKEN"))
	if err != nil {
		return err
	}

	var logWriter *os.File
	if auditLog != "" {
		logWriter, err = os.OpenFile(auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer logWriter.Close()
	}

	out := make(chan vault.VersionChange, 8)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	opts := vault.WatchOptions{Interval: interval}
	go vault.WatchSecret(ctx, c, path, opts, out)

	fmt.Fprintf(os.Stderr, "Watching %s every %s (Ctrl+C to stop)...\n", path, interval)
	for {
		select {
		case change := <-out:
			fmt.Printf("[%s] %s: v%d -> v%d\n",
				change.DetectedAt.Format(time.RFC3339), change.Path,
				change.FromVersion, change.ToVersion)
			if logWriter != nil {
				audit.LogWatchEvent(logWriter, change.Path, change.FromVersion, change.ToVersion)
			}
		case <-ctx.Done():
			return nil
		}
	}
}
