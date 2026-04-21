package vault

import "fmt"

// TagDiff represents a change in a secret tag.
type TagDiff struct {
	Key    string
	OldVal string
	NewVal string
	Status string // "added", "removed", "modified", "unchanged"
}

// DiffTags compares two tag maps and returns a list of TagDiff entries.
func DiffTags(oldTags, newTags map[string]string) []TagDiff {
	seen := map[string]bool{}
	var diffs []TagDiff

	for k, oldVal := range oldTags {
		seen[k] = true
		if newVal, ok := newTags[k]; ok {
			if oldVal != newVal {
				diffs = append(diffs, TagDiff{Key: k, OldVal: oldVal, NewVal: newVal, Status: "modified"})
			} else {
				diffs = append(diffs, TagDiff{Key: k, OldVal: oldVal, NewVal: newVal, Status: "unchanged"})
			}
		} else {
			diffs = append(diffs, TagDiff{Key: k, OldVal: oldVal, NewVal: "", Status: "removed"})
		}
	}

	for k, newVal := range newTags {
		if !seen[k] {
			diffs = append(diffs, TagDiff{Key: k, OldVal: "", NewVal: newVal, Status: "added"})
		}
	}

	return diffs
}

// FormatTagDiff returns a human-readable string of tag differences.
func FormatTagDiff(diffs []TagDiff) string {
	if len(diffs) == 0 {
		return "(no tag changes)"
	}
	out := ""
	for _, d := range diffs {
		switch d.Status {
		case "added":
			out += fmt.Sprintf("+ [%s] = %q\n", d.Key, d.NewVal)
		case "removed":
			out += fmt.Sprintf("- [%s] = %q\n", d.Key, d.OldVal)
		case "modified":
			out += fmt.Sprintf("~ [%s]: %q -> %q\n", d.Key, d.OldVal, d.NewVal)
		case "unchanged":
			out += fmt.Sprintf("  [%s] = %q\n", d.Key, d.NewVal)
		}
	}
	return out
}
