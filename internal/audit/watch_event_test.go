package audit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogWatchEvent_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	err := LogWatchEvent(&buf, "secret/data/myapp", 2, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var event WatchEvent
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if event.Type != "watch" {
		t.Errorf("expected type 'watch', got %q", event.Type)
	}
	if event.Path != "secret/data/myapp" {
		t.Errorf("unexpected path: %s", event.Path)
	}
	if event.FromVersion != 2 || event.ToVersion != 3 {
		t.Errorf("expected 2->3, got %d->%d", event.FromVersion, event.ToVersion)
	}
	if event.DetectedAt.IsZero() {
		t.Error("DetectedAt should not be zero")
	}
}

func TestLogWatchEvent_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 3; i++ {
		if err := LogWatchEvent(&buf, "secret/data/svc", i, i+1); err != nil {
			t.Fatalf("error on entry %d: %v", i, err)
		}
	}

	dec := json.NewDecoder(&buf)
	count := 0
	for dec.More() {
		var e WatchEvent
		if err := dec.Decode(&e); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}
}
