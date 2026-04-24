package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/yourusername/vaultdiff/internal/vault"
)

// AccessLogAuditEntry is the structured audit record for a secret access log query.
type AccessLogAuditEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Event      string    `json:"event"`
	Path       string    `json:"path"`
	EntryCount int       `json:"entry_count"`
	Summary    string    `json:"summary"`
}

// LogAccessLogEvent writes an audit entry for a secret access log retrieval.
func LogAccessLogEvent(w io.Writer, result *vault.AccessLogResult) error {
	if result == nil {
		return fmt.Errorf("nil access log result")
	}

	entry := AccessLogAuditEntry{
		Timestamp:  time.Now().UTC(),
		Event:      "access_log_read",
		Path:       result.Path,
		EntryCount: len(result.Entries),
		Summary:    summarizeAccessLog(result),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling access log audit entry: %w", err)
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}

func summarizeAccessLog(result *vault.AccessLogResult) string {
	if len(result.Entries) == 0 {
		return fmt.Sprintf("no access log entries found for %s", result.Path)
	}
	return fmt.Sprintf("%d version(s) recorded for secret %s", len(result.Entries), result.Path)
}
