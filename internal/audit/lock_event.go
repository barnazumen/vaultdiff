package audit

import (
	"encoding/json"
	"io"
	"time"
)

// LockEventEntry represents an audit log entry for a secret lock or unlock action.
type LockEventEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // "lock" or "unlock"
	Path      string    `json:"path"`
	Mount     string    `json:"mount"`
	LockedBy  string    `json:"locked_by,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	TTL       int       `json:"ttl_seconds,omitempty"`
}

// LogLockEvent writes a lock or unlock audit event to the provided writer.
func LogLockEvent(w io.Writer, action, path, mount, lockedBy, reason string, ttl int) error {
	entry := LockEventEntry{
		Timestamp: time.Now().UTC(),
		Action:    action,
		Path:      path,
		Mount:     mount,
		LockedBy:  lockedBy,
		Reason:    reason,
		TTL:       ttl,
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(entry)
}
