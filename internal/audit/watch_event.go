package audit

import (
	"encoding/json"
	"io"
	"time"
)

// WatchEvent records a version change detected during secret watching.
type WatchEvent struct {
	Type        string    `json:"type"`
	Path        string    `json:"path"`
	FromVersion int       `json:"from_version"`
	ToVersion   int       `json:"to_version"`
	DetectedAt  time.Time `json:"detected_at"`
}

// LogWatchEvent writes a WatchEvent as a JSON line to the given writer.
func LogWatchEvent(w io.Writer, path string, from, to int) error {
	event := WatchEvent{
		Type:        "watch",
		Path:        path,
		FromVersion: from,
		ToVersion:   to,
		DetectedAt:  time.Now().UTC(),
	}
	enc := json.NewEncoder(w)
	return enc.Encode(event)
}
