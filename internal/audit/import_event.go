package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ImportAuditEntry records the outcome of a bulk secret import operation.
type ImportAuditEntry struct {
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
	Mount     string `json:"mount"`
	File      string `json:"file"`
	Total     int    `json:"total"`
	Succeeded int    `json:"succeeded"`
	Failed    int    `json:"failed"`
	Paths     []string `json:"paths"`
}

// ImportResult mirrors the vault package type to avoid a circular dependency.
type ImportResult struct {
	Path    string
	Success bool
	Error   string
}

// LogImportEvent writes a structured JSON audit entry for an import operation.
func LogImportEvent(w io.Writer, mount, file string, results []ImportResult) error {
	var succeeded, failed int
	var paths []string
	for _, r := range results {
		paths = append(paths, r.Path)
		if r.Success {
			succeeded++
		} else {
			failed++
		}
	}

	entry := ImportAuditEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Event:     "secret_import",
		Mount:     mount,
		File:      file,
		Total:     len(results),
		Succeeded: succeeded,
		Failed:    failed,
		Paths:     paths,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal import audit entry: %w", err)
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}
