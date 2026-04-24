package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// CloneEvent represents an audit log entry for a secret clone operation.
type CloneEvent struct {
	Timestamp  string   `json:"timestamp"`
	Event      string   `json:"event"`
	SourcePath string   `json:"source_path"`
	DestPath   string   `json:"dest_path"`
	Mount      string   `json:"mount"`
	Version    int      `json:"version"`
	Keys       []string `json:"keys"`
	KeyCount   int      `json:"key_count"`
}

// LogCloneEvent writes a structured JSON audit entry for a secret clone.
func LogCloneEvent(w io.Writer, sourcePath, destPath, mount string, version int, keys []string) error {
	if mount == "" {
		mount = "secret"
	}
	entry := CloneEvent{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Event:      "secret_cloned",
		SourcePath: sourcePath,
		DestPath:   destPath,
		Mount:      mount,
		Version:    version,
		Keys:       keys,
		KeyCount:   len(keys),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("log clone event: marshal: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
