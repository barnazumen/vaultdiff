package main

import (
	"bytes"
	"testing"
)

func TestMoveCmd_MissingArgs(t *testing.T) {
	cmd := secretMoveCmd
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
}

func TestMoveCmd_MissingEnv(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	cmd := secretMoveCmd
	cmd.SetArgs([]string{"secret", "src/key", "dst/key"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{"secret", "src/key", "dst/key"})
	if err == nil {
		t.Fatal("expected error when VAULT_ADDR is missing")
	}
}

func TestMoveCmd_HelpOutput(t *testing.T) {
	var buf bytes.Buffer
	secretMoveCmd.SetOut(&buf)
	secretMoveCmd.SetArgs([]string{"--help"})

	// help should not return an error
	_ = secretMoveCmd.Help()

	if buf.Len() == 0 {
		t.Error("expected help output, got empty")
	}
}
