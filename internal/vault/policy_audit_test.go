package vault

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLogPolicyAudit_WriteEntry(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "policy_audit_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	diff := []PolicyLineDiff{
		{Type: "added", Line: "path \"secret/*\" { capabilities = [\"read\"] }"},
		{Type: "removed", Line: "path \"secret/*\" { capabilities = [\"list\"] }"},
		{Type: "unchanged", Line: "# comment"},
	}

	err = LogPolicyAudit(tmp.Name(), "policy-old", "policy-new", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tmp.Name())
	var entry PolicyAuditEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if entry.PolicyA != "policy-old" {
		t.Errorf("expected policy_a=policy-old, got %s", entry.PolicyA)
	}
	if !entry.HasDiff {
		t.Error("expected has_diff=true")
	}
	if len(entry.Added) != 1 || len(entry.Removed) != 1 || len(entry.Unchanged) != 1 {
		t.Errorf("unexpected counts: added=%d removed=%d unchanged=%d", len(entry.Added), len(entry.Removed), len(entry.Unchanged))
	}
}

func TestLogPolicyAudit_NoDiff(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "policy_audit_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	diff := []PolicyLineDiff{
		{Type: "unchanged", Line: "# same"},
	}

	err = LogPolicyAudit(tmp.Name(), "pol-a", "pol-b", diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(tmp.Name())
	var entry PolicyAuditEntry
	json.Unmarshal(data, &entry)

	if entry.HasDiff {
		t.Error("expected has_diff=false")
	}
}
