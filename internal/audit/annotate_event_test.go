package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogAnnotateEvent_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := LogAnnotateEvent(&buf, "secret", "myapp/config", "owner", "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var event AnnotateEvent
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &event); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if event.Action != "annotate" {
		t.Errorf("expected action=annotate, got %q", event.Action)
	}
	if event.Mount != "secret" {
		t.Errorf("expected mount=secret, got %q", event.Mount)
	}
	if event.Path != "myapp/config" {
		t.Errorf("expected path=myapp/config, got %q", event.Path)
	}
	if event.Key != "owner" {
		t.Errorf("expected key=owner, got %q", event.Key)
	}
	if event.Value != "alice" {
		t.Errorf("expected value=alice, got %q", event.Value)
	}
	if event.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestLogAnnotateEvent_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	for _, kv := range [][2]string{{"owner", "alice"}, {"team", "platform"}, {"env", "prod"}} {
		if err := LogAnnotateEvent(&buf, "secret", "myapp/config", kv[0], kv[1]); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 log lines, got %d", len(lines))
	}

	for i, line := range lines {
		var e AnnotateEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Errorf("line %d invalid JSON: %v", i, err)
		}
	}
}
