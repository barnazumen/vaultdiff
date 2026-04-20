package vault

import "testing"

func TestDiffMounts_Added(t *testing.T) {
	old := map[string]string{
		"secret/": "kv",
	}
	new_ := map[string]string{
		"secret/": "kv",
		"pki/":    "pki",
	}
	result := DiffMounts(old, new_)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result))
	}
	if result[0].Path != "pki/" || result[0].ChangeType != "added" {
		t.Errorf("unexpected diff: %+v", result[0])
	}
}

func TestDiffMounts_Removed(t *testing.T) {
	old := map[string]string{
		"secret/": "kv",
		"transit/": "transit",
	}
	new_ := map[string]string{
		"secret/": "kv",
	}
	result := DiffMounts(old, new_)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result))
	}
	if result[0].Path != "transit/" || result[0].ChangeType != "removed" {
		t.Errorf("unexpected diff: %+v", result[0])
	}
}

func TestDiffMounts_TypeChanged(t *testing.T) {
	old := map[string]string{"secret/": "kv"}
	new_ := map[string]string{"secret/": "kv-v2"}
	result := DiffMounts(old, new_)
	if len(result) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(result))
	}
	if result[0].ChangeType != "modified" {
		t.Errorf("expected modified, got %s", result[0].ChangeType)
	}
}

func TestDiffMounts_NoChange(t *testing.T) {
	old := map[string]string{"secret/": "kv"}
	result := DiffMounts(old, old)
	if len(result) != 0 {
		t.Fatalf("expected 0 diffs, got %d", len(result))
	}
}

func TestFormatMountDiff_Output(t *testing.T) {
	diffs := []MountDiff{
		{Path: "pki/", OldType: "", NewType: "pki", ChangeType: "added"},
		{Path: "old/", OldType: "kv", NewType: "", ChangeType: "removed"},
	}
	out := FormatMountDiff(diffs)
	if out == "" {
		t.Error("expected non-empty output")
	}
	for _, d := range diffs {
		if !contains(out, d.Path) {
			t.Errorf("expected path %s in output", d.Path)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
