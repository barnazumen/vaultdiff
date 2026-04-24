package audit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogPromoteEvent_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	err := LogPromoteEvent(&buf, "secret", "prod-secret", "myapp/config", 3, []string{"api_key", "env"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry PromoteEvent
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if entry.Event != "secret_promoted" {
		t.Errorf("expected event 'secret_promoted', got %q", entry.Event)
	}
	if entry.SourceMount != "secret" {
		t.Errorf("expected source_mount 'secret', got %q", entry.SourceMount)
	}
	if entry.DestMount != "prod-secret" {
		t.Errorf("expected dest_mount 'prod-secret', got %q", entry.DestMount)
	}
	if entry.Path != "myapp/config" {
		t.Errorf("expected path 'myapp/config', got %q", entry.Path)
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

func TestLogPromoteEvent_EmptyKeys(t *testing.T) {
	var buf bytes.Buffer
	err := LogPromoteEvent(&buf, "secret", "prod", "empty/path", 1, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry PromoteEvent
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.KeyCount != 0 {
		t.Errorf("expected key_count 0, got %d", entry.KeyCount)
	}
}
