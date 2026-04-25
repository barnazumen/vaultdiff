package vault

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/your/vaultdiff/internal/diff"
)

func sampleChanges() []diff.Change {
	return []diff.Change{
		{Key: "username", Type: diff.Unchanged, OldValue: "admin", NewValue: "admin"},
		{Key: "password", Type: diff.Modified, OldValue: "old", NewValue: "new"},
		{Key: "api_key", Type: diff.Added, OldValue: "", NewValue: "abc123"},
		{Key: "deprecated", Type: diff.Removed, OldValue: "legacy", NewValue: ""},
	}
}

func TestBuildDiffReport_Summary(t *testing.T) {
	changes := sampleChanges()
	report := BuildDiffReport("secret/myapp/config", 2, 3, changes)

	if report.Path != "secret/myapp/config" {
		t.Errorf("expected path 'secret/myapp/config', got %q", report.Path)
	}
	if report.FromVersion != 2 || report.ToVersion != 3 {
		t.Errorf("unexpected versions: from=%d to=%d", report.FromVersion, report.ToVersion)
	}
	if report.Summary.Added != 1 {
		t.Errorf("expected 1 added, got %d", report.Summary.Added)
	}
	if report.Summary.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", report.Summary.Removed)
	}
	if report.Summary.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", report.Summary.Modified)
	}
	if report.Summary.Unchanged != 1 {
		t.Errorf("expected 1 unchanged, got %d", report.Summary.Unchanged)
	}
	if len(report.Changes) != 4 {
		t.Errorf("expected 4 changes, got %d", len(report.Changes))
	}
}

func TestWriteReportJSON_CreatesValidFile(t *testing.T) {
	changes := sampleChanges()
	report := BuildDiffReport("secret/test", 1, 2, changes)

	tmpFile, err := os.CreateTemp("", "report-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := WriteReportJSON(report, tmpFile.Name()); err != nil {
		t.Fatalf("WriteReportJSON failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read report file: %v", err)
	}

	var decoded DiffReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode report JSON: %v", err)
	}
	if decoded.Path != "secret/test" {
		t.Errorf("expected path 'secret/test', got %q", decoded.Path)
	}
	if decoded.Summary.Modified != 1 {
		t.Errorf("expected 1 modified in decoded report, got %d", decoded.Summary.Modified)
	}
}

func TestBuildDiffReport_EmptyChanges(t *testing.T) {
	report := BuildDiffReport("secret/empty", 1, 1, []diff.Change{})
	if report.Summary.Added != 0 || report.Summary.Removed != 0 {
		t.Error("expected all zero counts for empty changes")
	}
	if !report.Timestamp.IsZero() == false {
		t.Error("expected non-zero timestamp")
	}
}
