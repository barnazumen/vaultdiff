package diff

import (
	"testing"
)

func TestCompare_Added(t *testing.T) {
	old := map[string]interface{}{}
	new_ := map[string]interface{}{"token": "abc"}

	r := Compare(old, new_)
	if len(r.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(r.Changes))
	}
	if r.Changes[0].Type != Added {
		t.Errorf("expected Added, got %s", r.Changes[0].Type)
	}
	if r.Changes[0].NewValue != "abc" {
		t.Errorf("unexpected new value: %s", r.Changes[0].NewValue)
	}
}

func TestCompare_Removed(t *testing.T) {
	old := map[string]interface{}{"token": "abc"}
	new_ := map[string]interface{}{}

	r := Compare(old, new_)
	if r.Changes[0].Type != Removed {
		t.Errorf("expected Removed, got %s", r.Changes[0].Type)
	}
}

func TestCompare_Modified(t *testing.T) {
	old := map[string]interface{}{"password": "old"}
	new_ := map[string]interface{}{"password": "new"}

	r := Compare(old, new_)
	if r.Changes[0].Type != Modified {
		t.Errorf("expected Modified, got %s", r.Changes[0].Type)
	}
	if r.Changes[0].OldValue != "old" || r.Changes[0].NewValue != "new" {
		t.Errorf("unexpected values: old=%s new=%s", r.Changes[0].OldValue, r.Changes[0].NewValue)
	}
}

func TestCompare_Unchanged(t *testing.T) {
	old := map[string]interface{}{"key": "val"}
	new_ := map[string]interface{}{"key": "val"}

	r := Compare(old, new_)
	if r.Changes[0].Type != Unchanged {
		t.Errorf("expected Unchanged, got %s", r.Changes[0].Type)
	}
	if r.HasChanges() {
		t.Error("expected HasChanges to be false")
	}
}

func TestCompare_Mixed(t *testing.T) {
	old := map[string]interface{}{"a": "1", "b": "2"}
	new_ := map[string]interface{}{"a": "1", "c": "3"}

	r := Compare(old, new_)
	if !r.HasChanges() {
		t.Error("expected changes")
	}
	typeMap := map[string]ChangeType{}
	for _, ch := range r.Changes {
		typeMap[ch.Key] = ch.Type
	}
	if typeMap["a"] != Unchanged {
		t.Errorf("a should be unchanged")
	}
	if typeMap["b"] != Removed {
		t.Errorf("b should be removed")
	}
	if typeMap["c"] != Added {
		t.Errorf("c should be added")
	}
}
