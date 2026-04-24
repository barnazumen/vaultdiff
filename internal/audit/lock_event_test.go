package audit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogLockEvent_Lock_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	err := LogLockEvent(&buf, "lock", "myapp/db", "secret", "tok-abc", "maintenance", 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var entry LockEventEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Action != "lock" {
		t.Errorf("expected action 'lock', got %q", entry.Action)
	}
	if entry.Path != "myapp/db" {
		t.Errorf("expected path 'myapp/db', got %q", entry.Path)
	}
	if entry.Mount != "secret" {
		t.Errorf("expected mount 'secret', got %q", entry.Mount)
	}
	if entry.Reason != "maintenance" {
		t.Errorf("expected reason 'maintenance', got %q", entry.Reason)
	}
	if entry.TTL != 3600 {
		t.Errorf("expected TTL 3600, got %d", entry.TTL)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestLogLockEvent_Unlock_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	err := LogLockEvent(&buf, "unlock", "myapp/db", "secret", "tok-abc", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var entry LockEventEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Action != "unlock" {
		t.Errorf("expected action 'unlock', got %q", entry.Action)
	}
	if entry.TTL != 0 {
		t.Errorf("expected TTL 0 for unlock, got %d", entry.TTL)
	}
}

func TestLogLockEvent_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 3; i++ {
		if err := LogLockEvent(&buf, "lock", "path", "secret", "tok", "reason", 60); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}
	}
	dec := json.NewDecoder(&buf)
	count := 0
	for dec.More() {
		var e LockEventEntry
		if err := dec.Decode(&e); err != nil {
			t.Fatalf("decode entry %d: %v", count, err)
		}
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}
}
