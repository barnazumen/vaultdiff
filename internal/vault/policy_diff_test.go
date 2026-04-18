package vault

import (
	"strings"
	"testing"
)

func TestDiffPolicies_Added(t *testing.T) {
	old := "path \"secret/*\" {\n  capabilities = [\"read\"]\n}"
	new_ := "path \"secret/*\" {\n  capabilities = [\"read\"]\n}\npath \"kv/*\" {\n  capabilities = [\"list\"]\n}"

	diff := DiffPolicies(old, new_)
	if len(diff.Added) != 2 {
		t.Errorf("expected 2 added lines, got %d", len(diff.Added))
	}
	if len(diff.Removed) != 0 {
		t.Errorf("expected 0 removed lines, got %d", len(diff.Removed))
	}
}

func TestDiffPolicies_Removed(t *testing.T) {
	old := "path \"secret/*\" {\n  capabilities = [\"read\"]\n}\npath \"kv/*\" {\n  capabilities = [\"list\"]\n}"
	new_ := "path \"secret/*\" {\n  capabilities = [\"read\"]\n}"

	diff := DiffPolicies(old, new_)
	if len(diff.Removed) != 2 {
		t.Errorf("expected 2 removed lines, got %d", len(diff.Removed))
	}
	if len(diff.Added) != 0 {
		t.Errorf("expected 0 added lines, got %d", len(diff.Added))
	}
}

func TestDiffPolicies_Unchanged(t *testing.T) {
	policy := "path \"secret/*\" {\n  capabilities = [\"read\"]\n}"
	diff := DiffPolicies(policy, policy)
	if len(diff.Added) != 0 || len(diff.Removed) != 0 {
		t.Errorf("expected no changes for identical policies")
	}
	if len(diff.Unchanged) == 0 {
		t.Errorf("expected unchanged lines")
	}
}

func TestDiffPolicies_Empty(t *testing.T) {
	diff := DiffPolicies("", "")
	if len(diff.Added) != 0 || len(diff.Removed) != 0 || len(diff.Unchanged) != 0 {
		t.Errorf("expected empty diff for empty policies")
	}
}

func TestFormatPolicyDiff_ContainsMarkers(t *testing.T) {
	diff := PolicyDiff{
		Added:   []string{"new line"},
		Removed: []string{"old line"},
		Unchanged: []string{"same line"},
	}
	out := FormatPolicyDiff(diff)
	if !strings.Contains(out, "+ new line") {
		t.Errorf("expected '+ new line' in output")
	}
	if !strings.Contains(out, "- old line") {
		t.Errorf("expected '- old line' in output")
	}
	if !strings.Contains(out, "  same line") {
		t.Errorf("expected '  same line' in output")
	}
}
