package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// MoveEvent records a secret move operation in the audit log.
type MoveEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Event       string    `json:"event"`
	Mount       string    `json:"mount"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Versions    int       `json:"versions_copied"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
}

// LogMoveEvent writes a move audit event as a JSON line to the provided writer.
func LogMoveEvent(w io.Writer, mount, src, dst string, versions int, opErr error) error {
	ev := MoveEvent{
		Timestamp:   time.Now().UTC(),
		Event:       "secret.move",
		Mount:       mount,
		Source:      src,
		Destination: dst,
		Versions:    versions,
		Success:     opErr == nil,
	}
	if opErr != nil {
		ev.Error = opErr.Error()
	}

	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("log move event: marshal: %w", err)
	}

	_, err = fmt.Fprintln(w, string(data))
	if err != nil {
		return fmt.Errorf("log move event: write: %w", err)
	}
	return nil
}
