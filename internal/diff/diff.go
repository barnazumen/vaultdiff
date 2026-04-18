package diff

import (
	"fmt"
	"sort"
	"strings"
)

// ChangeType represents the type of change for a secret key.
type ChangeType string

const (
	Added    ChangeType = "added"
	Removed  ChangeType = "removed"
	Modified ChangeType = "modified"
	Unchanged ChangeType = "unchanged"
)

// Change represents a single key-level diff entry.
type Change struct {
	Key      string
	Type     ChangeType
	OldValue string
	NewValue string
}

// Result holds the full diff between two secret versions.
type Result struct {
	Changes []Change
}

// HasChanges returns true if there are any non-unchanged entries.
func (r *Result) HasChanges() bool {
	for _, c := range r.Changes {
		if c.Type != Unchanged {
			return true
		}
	}
	return false
}

// Compare diffs two secret data maps and returns a Result.
func Compare(oldData, newData map[string]interface{}) *Result {
	keys := unionKeys(oldData, newData)
	sort.Strings(keys)

	var changes []Change
	for _, k := range keys {
		oldVal, inOld := stringify(oldData, k)
		newVal, inNew := stringify(newData, k)

		switch {
		case inOld && !inNew:
			changes = append(changes, Change{Key: k, Type: Removed, OldValue: oldVal})
		case !inOld && inNew:
			changes = append(changes, Change{Key: k, Type: Added, NewValue: newVal})
		case oldVal != newVal:
			changes = append(changes, Change{Key: k, Type: Modified, OldValue: oldVal, NewValue: newVal})
		default:
			changes = append(changes, Change{Key: k, Type: Unchanged, OldValue: oldVal, NewValue: newVal})
		}
	}
	return &Result{Changes: changes}
}

func unionKeys(a, b map[string]interface{}) []string {
	seen := map[string]struct{}{}
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	return keys
}

func stringify(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v)), true
}
