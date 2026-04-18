package audit

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/user/vaultdiff/internal/diff"
)

func TestLog_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	changes := []diff.Change{
		{Key: "foo", Type: diff.Added, NewValue: "bar"},
		{Key: "baz", Type: diff.Removed, OldValue: "qux"},
		{Key: "x", Type: diff.Modified, OldValue: "1", NewValue: "2"},
		{Key: "y", Type: diff.Unchanged, OldValue: "z", NewValue: "z"},
	}

	if err := logger.Log("secret/myapp", 1, 2, changes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if entry.Path != "secret/myapp" {
		t.Errorf("expected path secret/myapp, got %s", entry.Path)
	}
	if entry.VersionA != 1 || entry.VersionB != 2 {
		t.Errorf("unexpected versions: %d %d", entry.VersionA, entry.VersionB)
	}
	if len(entry.Changes) != 4 {
		t.Errorf("expected 4 changes, got %d", len(entry.Changes))
	}
}

func TestSummarize(t *testing.T) {
	changes := []diff.Change{
		{Type: diff.Added},
		{Type: diff.Added},
		{Type: diff.Removed},
		{Type: diff.Modified},
		{Type: diff.Unchanged},
	}
	s := summarize(changes)
	if s.Added != 2 || s.Removed != 1 || s.Modified != 1 || s.Unchanged != 1 {
		t.Errorf("unexpected summary: %+v", s)
	}
}

func TestLog_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	for i := 0; i < 3; i++ {
		if err := logger.Log("secret/test", i, i+1, nil); err != nil {
			t.Fatalf("log error: %v", err)
		}
	}

	lines := bytes.Split(bytes.TrimRight(buf.Bytes(), "\n"), []byte("\n"))
	if len(lines) != 3 {
		t.Errorf("expected 3 log lines, got %d", len(lines))
	}
}
