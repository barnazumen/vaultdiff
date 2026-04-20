package audit

import (
	"strings"

	"github.com/user/vaultdiff/internal/diff"
)

// FilterOptions controls which changes are included in an audit entry.
type FilterOptions struct {
	// OnlyTypes restricts changes to these types. Empty means all types.
	OnlyTypes []diff.ChangeType
	// KeyPrefix filters changes to keys with the given prefix.
	KeyPrefix string
}

// Filter returns a subset of changes matching the given options.
func Filter(changes []diff.Change, opts FilterOptions) []diff.Change {
	var result []diff.Change
	for _, c := range changes {
		if !matchesType(c.Type, opts.OnlyTypes) {
			continue
		}
		if opts.KeyPrefix != "" && !strings.HasPrefix(c.Key, opts.KeyPrefix) {
			continue
		}
		result = append(result, c)
	}
	return result
}

// FilterByTypes returns only changes whose type is in the provided list.
// It is a convenience wrapper around Filter for type-only filtering.
func FilterByTypes(changes []diff.Change, types ...diff.ChangeType) []diff.Change {
	return Filter(changes, FilterOptions{OnlyTypes: types})
}

func matchesType(ct diff.ChangeType, allowed []diff.ChangeType) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if a == ct {
			return true
		}
	}
	return false
}
