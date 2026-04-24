package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// AnnotateEvent records a secret annotation change in the audit log.
type AnnotateEvent struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Mount     string `json:"mount"`
	Path      string `json:"path"`
	Key       string `json:"annotation_key"`
	Value     string `json:"annotation_value"`
}

// LogAnnotateEvent writes an annotation audit entry to w.
func LogAnnotateEvent(w io.Writer, mount, path, key, value string) error {
	event := AnnotateEvent{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Action:    "annotate",
		Mount:     mount,
		Path:      path,
		Key:       key,
		Value:     value,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal annotate event: %w", err)
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}
