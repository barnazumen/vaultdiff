package vault

import "fmt"

// MountDiff describes a change between two sets of mounts.
type MountDiff struct {
	Path   string
	Status string // "added", "removed", "changed"
	Old    *MountInfo
	New    *MountInfo
}

// DiffMounts compares two mount maps and returns a list of differences.
func DiffMounts(before, after map[string]MountInfo) []MountDiff {
	var diffs []MountDiff

	for path, bInfo := range before {
		aInfo, exists := after[path]
		if !exists {
			b := bInfo
			diffs = append(diffs, MountDiff{Path: path, Status: "removed", Old: &b})
			continue
		}
		if bInfo.Type != aInfo.Type || bInfo.Description != aInfo.Description {
			b, a := bInfo, aInfo
			diffs = append(diffs, MountDiff{Path: path, Status: "changed", Old: &b, New: &a})
		}
	}

	for path, aInfo := range after {
		if _, exists := before[path]; !exists {
			a := aInfo
			diffs = append(diffs, MountDiff{Path: path, Status: "added", New: &a})
		}
	}
	return diffs
}

// FormatMountDiff returns a human-readable string for a MountDiff slice.
func FormatMountDiff(diffs []MountDiff) string {
	if len(diffs) == 0 {
		return "No mount changes detected."
	}
	out := ""
	for _, d := range diffs {
		switch d.Status {
		case "added":
			out += fmt.Sprintf("+ %s (type: %s)\n", d.Path, d.New.Type)
		case "removed":
			out += fmt.Sprintf("- %s (type: %s)\n", d.Path, d.Old.Type)
		case "changed":
			out += fmt.Sprintf("~ %s: %s -> %s\n", d.Path, d.Old.Type, d.New.Type)
		}
	}
	return out
}
