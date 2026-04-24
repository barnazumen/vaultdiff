package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// PromoteEvent represents an audit log entry for a secret promotion.
type PromoteEvent struct {
	Timestamp   string   `json:"timestamp"`
	Event       string   `json:"event"`
	SourceMount string   `json:"source_mount"`
	DestMount   string   `json:"dest_mount"`
	Path        string   `json:"path"`
	Version     int      `json:"version"`
	Keys        []string `json:"keys"`
	KeyCount    int      `json:"key_count"`
}

// LogPromoteEvent writes a structured JSON audit entry for a secret promotion.
func LogPromoteEvent(w io.Writer, srcMount, dstMount, path string, version int, keys []string) error {
	entry := PromoteEvent{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Event:       "secret_promoted",
		SourceMount: srcMount,
		DestMount:   dstMount,
		Path:        path,
		Version:     version,
		Keys:        keys,
		KeyCount:    len(keys),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("log promote event: marshal: %w", err)
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}
