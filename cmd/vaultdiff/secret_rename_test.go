package main

import (
	"bytes"
	"testing"
)

func TestRenameCmd_MissingArgs(t *testing.T) {
	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"rename"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args, got nil")
	}
}

func TestRenameCmd_MissingEnv(t *testing.T) {
	t.Setenv("VAULT_ADDR", "")
	t.Setenv("VAULT_TOKEN", "")

	err := runRenameSecret("old", "new", "secret")
	if err == nil {
		t.Fatal("expected error when env vars missing")
	}
}

func TestRenameCmd_HelpOutput(t *testing.T) {
	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{"rename", "--help"})

	_ = rootCmd.Execute()

	if !bytes.Contains(out.Bytes(), []byte("rename")) {
		t.Errorf("expected 'rename' in help output, got: %s", out.String())
	}
}
