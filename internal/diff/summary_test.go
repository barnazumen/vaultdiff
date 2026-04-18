package diff

import (
	"strings"
	"testing"
)

func makeSampleChanges() []ChangeRecord {
	return []ChangeRecord{
		{Key: "a", Type: "added", NewValue: "1"},
		{Key: "b", Type: "removed", OldValue: "2"},
		{Key: "c", Type: "modified", OldValue: "3", NewValue: "4"},
		{Key: "d", Type: "unchanged", OldValue: "5", NewValue: "5"},
		{Key: "e", Type: "added", NewValue: "6"},
	}
}

func TestSummarize_Counts(t *testing.T) {
	changes := makeSampleChanges()
	s := Summarize(changes)

	if s.Added != 2 {
		t.Errorf("expected Added=2, got %d", s.Added)
	}
	if s.Removed != 1 {
		t.Errorf("expected Removed=1, got %d", s.Removed)
	}
	if s.Modified != 1 {
		t.Errorf("expected Modified=1, got %d", s.Modified)
	}
	if s.Unchanged != 1 {
		t.Errorf("expected Unchanged=1, got %d", s.Unchanged)
	}
	if s.Total != 5 {
		t.Errorf("expected Total=5, got %d", s.Total)
	}
}

func TestSummarize_Empty(t *testing.T) {
	s := Summarize([]ChangeRecord{})
	if s.Total != 0 {
		t.Errorf("expected Total=0, got %d", s.Total)
	}
}

func TestSummary_String(t *testing.T) {
	s := Summary{Added: 2, Removed: 1, Modified: 1, Unchanged: 1, Total: 5}
	out := s.String()
	for _, want := range []string{"2 added", "1 removed", "1 modified", "1 unchanged", "total: 5"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in summary string, got: %s", want, out)
		}
	}
}
