package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// PinEvent records a pin or unpin operation against a secret version.
type PinEvent struct {
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
	Path      string `json:"path"`
	Mount     string `json:"mount"`
	Version   int    `json:"version"`
	Pinned    bool   `json:"pinned"`
	Actor     string `json:"actor,omitempty"`
}

// LogPinEvent writes a structured JSON pin/unpin audit entry to w.
func LogPinEvent(w io.Writer, mount, path string, version int, pinned bool, actor string) error {
	eventType := "secret.pin"
	if !pinned {
		eventType = "secret.unpin"
	}

	entry := PinEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     eventType,
		Path:      path,
		Mount:     mount,
		Version:   version,
		Pinned:    pinned,
		Actor:     actor,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("log pin event: marshal: %w", err)
	}

	_, err = fmt.Fprintln(w, string(b))
	return err
}
