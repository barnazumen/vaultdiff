package audit

import (
	"encoding/json"
	"io"
	"time"

	"github.com/user/vaultdiff/internal/diff"
)

// Entry represents a single audit log record for a diff operation.
type Entry struct {
	Timestamp time.Time         `json:"timestamp"`
	Path      string            `json:"path"`
	VersionA  int               `json:"version_a"`
	VersionB  int               `json:"version_b"`
	Changes   []diff.Change     `json:"changes"`
	Summary   Summary           `json:"summary"`
}

// Summary holds counts of change types.
type Summary struct {
	Added     int `json:"added"`
	Removed   int `json:"removed"`
	Modified  int `json:"modified"`
	Unchanged int `json:"unchanged"`
}

// Logger writes audit entries as newline-delimited JSON.
type Logger struct {
	w io.Writer
}

// NewLogger creates a Logger that writes to w.
func NewLogger(w io.Writer) *Logger {
	return &Logger{w: w}
}

// Log writes an audit entry derived from the given diff changes.
func (l *Logger) Log(path string, versionA, versionB int, changes []diff.Change) error {
	entry := Entry{
		Timestamp: time.Now().UTC(),
		Path:      path,
		VersionA:  versionA,
		VersionB:  versionB,
		Changes:   changes,
		Summary:   summarize(changes),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = l.w.Write(append(data, '\n'))
	return err
}

func summarize(changes []diff.Change) Summary {
	var s Summary
	for _, c := range changes {
		switch c.Type {
		case diff.Added:
			s.Added++
		case diff.Removed:
			s.Removed++
		case diff.Modified:
			s.Modified++
		case diff.Unchanged:
			s.Unchanged++
		}
	}
	return s
}
