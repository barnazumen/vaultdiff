package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// AuditTrailEventEntry is the log record written for an audit trail fetch.
type AuditTrailEventEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Event      string    `json:"event"`
	Path       string    `json:"path"`
	Mount      string    `json:"mount"`
	EntryCount int       `json:"entry_count"`
	Summary    string    `json:"summary"`
}

// LogAuditTrailEvent writes a structured JSON log entry for an audit trail read.
func LogAuditTrailEvent(w io.Writer, mount, path string, entryCount int) error {
	entry := AuditTrailEventEntry{
		Timestamp:  time.Now().UTC(),
		Event:      "audit_trail_read",
		Path:       path,
		Mount:      mount,
		EntryCount: entryCount,
		Summary:    summarizeAuditTrail(entryCount),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling audit trail event: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func summarizeAuditTrail(count int) string {
	if count == 0 {
		return "no audit trail entries found"
	}
	if count == 1 {
		return "1 audit trail entry retrieved"
	}
	return fmt.Sprintf("%d audit trail entries retrieved", count)
}
