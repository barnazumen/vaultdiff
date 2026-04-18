package audit

import (
	"testing"

	"github.com/user/vaultdiff/internal/diff"
)

var testChanges = []diff.Change{
	{Key: "db_pass", Type: diff.Added, NewValue: "secret"},
	{Key: "db_host", Type: diff.Unchanged, OldValue: "localhost", NewValue: "localhost"},
	{Key: "api_key", Type: diff.Modified, OldValue: "old", NewValue: "new"},
	{Key: "db_port", Type: diff.Removed, OldValue: "5432"},
}

func TestFilter_ByType(t *testing.T) {
	result := Filter(testChanges, FilterOptions{
		OnlyTypes: []diff.ChangeType{diff.Added, diff.Removed},
	})
	if len(result) != 2 {
		t.Errorf("expected 2 changes, got %d", len(result))
	}
}

func TestFilter_ByPrefix(t *testing.T) {
	result := Filter(testChanges, FilterOptions{
		KeyPrefix: "db_",
	})
	if len(result) != 3 {
		t.Errorf("expected 3 changes with prefix db_, got %d", len(result))
	}
}

func TestFilter_ByTypeAndPrefix(t *testing.T) {
	result := Filter(testChanges, FilterOptions{
		OnlyTypes: []diff.ChangeType{diff.Added},
		KeyPrefix: "db_",
	})
	if len(result) != 1 || result[0].Key != "db_pass" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestFilter_NoOptions(t *testing.T) {
	result := Filter(testChanges, FilterOptions{})
	if len(result) != len(testChanges) {
		t.Errorf("expected all %d changes, got %d", len(testChanges), len(result))
	}
}

func TestFilter_EmptyInput(t *testing.T) {
	result := Filter(nil, FilterOptions{OnlyTypes: []diff.ChangeType{diff.Added}})
	if result != nil {
		t.Errorf("expected nil result for empty input")
	}
}
