package audit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogTagEvent_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	err := LogTagEvent(&buf, "secret/myapp/config",
		map[string]string{"env": "prod"},
		map[string]string{"deprecated": "true"},
		map[string]string{"owner": "team-b"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var entry TagAuditEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Path != "secret/myapp/config" {
		t.Errorf("expected path 'secret/myapp/config', got %q", entry.Path)
	}
	if entry.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestLogTagEvent_NoChanges(t *testing.T) {
	var buf bytes.Buffer
	err := LogTagEvent(&buf, "secret/empty", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var entry TagAuditEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.Summary != "no changes" {
		t.Errorf("expected 'no changes', got %q", entry.Summary)
	}
}

func TestSummarizeTagChanges(t *testing.T) {
	cases := []struct {
		added, removed, modified map[string]string
		expect string
	}{
		{map[string]string{"a": "1"}, nil, nil, "1 added"},
		{nil, map[string]string{"b": "2"}, nil, "1 removed"},
		{nil, nil, map[string]string{"c": "3"}, "1 modified"},
		{nil, nil, nil, "no changes"},
	}
	for _, c := range cases {
		got := summarizeTagChanges(c.added, c.removed, c.modified)
		if got != c.expect {
			t.Errorf("expected %q, got %q", c.expect, got)
		}
	}
}
