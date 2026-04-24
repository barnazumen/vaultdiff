package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ValidateEvent represents an audit log entry for a secret validation check.
type ValidateEvent struct {
	Timestamp  string   `json:"timestamp"`
	Event      string   `json:"event"`
	Path       string   `json:"path"`
	Version    int      `json:"version"`
	Valid      bool     `json:"valid"`
	Missing    []string `json:"missing_keys,omitempty"`
	Extra      []string `json:"extra_keys,omitempty"`
	StrictMode bool     `json:"strict_mode"`
}

// LogValidateEvent writes a validation audit entry to the provided writer.
func LogValidateEvent(w io.Writer, path string, version int, valid bool, missing, extra []string, strict bool) error {
	entry := ValidateEvent{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Event:      "secret_validate",
		Path:       path,
		Version:    version,
		Valid:      valid,
		Missing:    missing,
		Extra:      extra,
		StrictMode: strict,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling validate event: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
