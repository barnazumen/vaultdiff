package vault

import (
	"strings"
	"testing"
)

func TestDiffTags_Added(t *testing.T) {
	old := map[string]string{}
	new_ := map[string]string{"env": "prod"}
	diffs := DiffTags(old, new_)
	if len(diffs) != 1 || diffs[0].Status != "added" || diffs[0].Key != "env" {
		t.Errorf("expected 1 added diff, got %+v", diffs)
	}
}

func TestDiffTags_Removed(t *testing.T) {
	old := map[string]string{"env": "prod"}
	new_ := map[string]string{}
	diffs := DiffTags(old, new_)
	if len(diffs) != 1 || diffs[0].Status != "removed" {
		t.Errorf("expected 1 removed diff, got %+v", diffs)
	}
}

func TestDiffTags_Modified(t *testing.T) {
	old := map[string]string{"env": "staging"}
	new_ := map[string]string{"env": "prod"}
	diffs := DiffTags(old, new_)
	if len(diffs) != 1 || diffs[0].Status != "modified" {
		t.Errorf("expected 1 modified diff, got %+v", diffs)
	}
	if diffs[0].OldVal != "staging" || diffs[0].NewVal != "prod" {
		t.Errorf("unexpected values: %+v", diffs[0])
	}
}

func TestDiffTags_Unchanged(t *testing.T) {
	old := map[string]string{"team": "infra"}
	new_ := map[string]string{"team": "infra"}
	diffs := DiffTags(old, new_)
	if len(diffs) != 1 || diffs[0].Status != "unchanged" {
		t.Errorf("expected 1 unchanged diff, got %+v", diffs)
	}
}

func TestFormatTagDiff_ContainsMarkers(t *testing.T) {
	old := map[string]string{"a": "1", "b": "old"}
	new_ := map[string]string{"b": "new", "c": "3"}
	diffs := DiffTags(old, new_)
	out := FormatTagDiff(diffs)
	if !strings.Contains(out, "- ") && !strings.Contains(out, "+ ") {
		t.Errorf("expected diff markers in output, got: %s", out)
	}
}

func TestFormatTagDiff_Empty(t *testing.T) {
	out := FormatTagDiff([]TagDiff{})
	if out != "(no tag changes)" {
		t.Errorf("expected empty message, got: %s", out)
	}
}
