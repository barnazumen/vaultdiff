package diff

import "fmt"

// Summary holds aggregated counts of diff changes.
type Summary struct {
	Added     int
	Removed   int
	Modified  int
	Unchanged int
	Total     int
}

// Summarize computes a Summary from a slice of ChangeRecords.
func Summarize(changes []ChangeRecord) Summary {
	var s Summary
	for _, c := range changes {
		switch c.Type {
		case "added":
			s.Added++
		case "removed":
			s.Removed++
		case "modified":
			s.Modified++
		case "unchanged":
			s.Unchanged++
		}
		s.Total++
	}
	return s
}

// String returns a human-readable one-line summary.
func (s Summary) String() string {
	return fmt.Sprintf(
		"Summary: %d added, %d removed, %d modified, %d unchanged (total: %d)",
		s.Added, s.Removed, s.Modified, s.Unchanged, s.Total,
	)
}
