package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// TagAuditEntry records a tag diff audit event.
type TagAuditEntry struct {
	Timestamp    string            `json:"timestamp"`
	Path         string            `json:"path"`
	AddedTags    map[string]string `json:"added_tags,omitempty"`
	RemovedTags  map[string]string `json:"removed_tags,omitempty"`
	ModifiedTags map[string]string `json:"modified_tags,omitempty"`
	Summary      string            `json:"summary"`
}

// LogTagEvent writes a tag audit entry to the provided writer.
func LogTagEvent(w io.Writer, path string, added, removed, modified map[string]string) error {
	summary := summarizeTagChanges(added, removed, modified)
	entry := TagAuditEntry{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Path:         path,
		AddedTags:    added,
		RemovedTags:  removed,
		ModifiedTags: modified,
		Summary:      summary,
	}
	return json.NewEncoder(w).Encode(entry)
}

func summarizeTagChanges(added, removed, modified map[string]string) string {
	parts := ""
	if len(added) > 0 {
		parts += fmt.Sprintf("%d added", len(added))
	}
	if len(removed) > 0 {
		if parts != "" {
			parts += ", "
		}
		parts += fmt.Sprintf("%d removed", len(removed))
	}
	if len(modified) > 0 {
		if parts != "" {
			parts += ", "
		}
		parts += fmt.Sprintf("%d modified", len(modified))
	}
	if parts == "" {
		return "no changes"
	}
	return parts
}
