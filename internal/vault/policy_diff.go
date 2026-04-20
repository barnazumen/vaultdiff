package vault

import (
	"fmt"
	"strings"
)

// PolicyDiff represents a line-level diff between two policy versions.
type PolicyDiff struct {
	Added   []string
	Removed []string
	Unchanged []string
}

// HasChanges reports whether the diff contains any added or removed lines.
func (d PolicyDiff) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0
}

// DiffPolicies compares two policy strings and returns a PolicyDiff.
func DiffPolicies(oldPolicy, newPolicy string) PolicyDiff {
	oldLines := splitLines(oldPolicy)
	newLines := splitLines(newPolicy)

	oldSet := toSet(oldLines)
	newSet := toSet(newLines)

	var diff PolicyDiff

	for _, line := range newLines {
		if _, exists := oldSet[line]; !exists {
			diff.Added = append(diff.Added, line)
		} else {
			diff.Unchanged = append(diff.Unchanged, line)
		}
	}

	for _, line := range oldLines {
		if _, exists := newSet[line]; !exists {
			diff.Removed = append(diff.Removed, line)
		}
	}

	return diff
}

// FormatPolicyDiff returns a human-readable unified-style diff string.
func FormatPolicyDiff(diff PolicyDiff) string {
	var sb strings.Builder
	for _, line := range diff.Removed {
		sb.WriteString(fmt.Sprintf("- %s\n", line))
	}
	for _, line := range diff.Added {
		sb.WriteString(fmt.Sprintf("+ %s\n", line))
	}
	for _, line := range diff.Unchanged {
		sb.WriteString(fmt.Sprintf("  %s\n", line))
	}
	return sb.String()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}

func toSet(lines []string) map[string]struct{} {
	m := make(map[string]struct{}, len(lines))
	for _, l := range lines {
		m[l] = struct{}{}
	}
	return m
}
