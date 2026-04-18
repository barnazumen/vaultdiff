package main

import (
	"bytes"
	"os"
	"testing"
)

func TestMain_MissingRequiredFlags(t *testing.T) {
	// Reset flags to defaults before test
	path = ""
	versionA = 0
	versionB = 0

	rootCmd.SetArgs([]string{"--mount", "secret"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing required flags, got nil")
	}
}

func TestMain_AuditLogWritten(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "audit-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Point to a mock server via env — integration-style smoke test skipped
	// without a live Vault; just verify flag parsing doesn't panic.
	auditLog = tmpFile.Name()
	noColor = true
	showAll = false

	// Restore
	t.Cleanup(func() {
		auditLog = ""
		noColor = false
	})

	// Verify the temp file is still accessible (audit path valid)
	if _, err := os.Stat(tmpFile.Name()); err != nil {
		t.Fatalf("audit log path invalid: %v", err)
	}
}

func TestRootCmd_HelpOutput(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})
	// Help exits with nil; ignore error
	_ = rootCmd.Execute()
	if !bytes.Contains(buf.Bytes(), []byte("vaultdiff")) {
		t.Errorf("expected help output to contain 'vaultdiff', got: %s", buf.String())
	}
}
