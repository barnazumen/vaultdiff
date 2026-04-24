package audit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogCloneEvent_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	err := LogCloneEvent(&buf, "src/mykey", "dst/mykey", "secret", 3, []string{"username", "password"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry CloneEvent
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Event != "secret_cloned" {
		t.Errorf("expected event secret_cloned, got %s", entry.Event)
	}
	if entry.SourcePath != "src/mykey" {
		t.Errorf("expected source src/mykey, got %s", entry.SourcePath)
	}
	if entry.DestPath != "dst/mykey" {
		t.Errorf("expected dest dst/mykey, got %s", entry.DestPath)
	}
	if entry.Version != 3 {
		t.Errorf("expected version 3, got %d", entry.Version)
	}
	if entry.KeyCount != 2 {
		t.Errorf("expected key_count 2, got %d", entry.KeyCount)
	}
	if entry.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestLogCloneEvent_DefaultMount(t *testing.T) {
	var buf bytes.Buffer
	err := LogCloneEvent(&buf, "a/b", "c/d", "", 1, []string{"key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry CloneEvent
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Mount != "secret" {
		t.Errorf("expected default mount 'secret', got %s", entry.Mount)
	}
}

func TestLogCloneEvent_EmptyKeys(t *testing.T) {
	var buf bytes.Buffer
	err := LogCloneEvent(&buf, "a/b", "c/d", "kv", 2, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry CloneEvent
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.KeyCount != 0 {
		t.Errorf("expected key_count 0, got %d", entry.KeyCount)
	}
}
