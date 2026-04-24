package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/yourusername/vaultdiff/internal/vault"
)

// ExpireEntry represents an audit log entry for a secret expiry check.
type ExpireEntry struct {
	Timestamp  string `json:"timestamp"`
	Event      string `json:"event"`
	Path       string `json:"path"`
	Version    int    `json:"version"`
	DaysOld    int    `json:"days_old"`
	MaxAgeDays int    `json:"max_age_days"`
	Expired    bool   `json:"expired"`
	Status     string `json:"status"`
}

// LogExpireEvent writes a structured JSON audit entry for a secret expiry check.
func LogExpireEvent(w io.Writer, expiry *vault.SecretExpiry, maxAgeDays int) error {
	status := "ok"
	if expiry.Expired {
		status = "expired"
	}

	entry := ExpireEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Event:      "secret_expiry_check",
		Path:       expiry.Path,
		Version:    expiry.Version,
		DaysOld:    expiry.DaysOld,
		MaxAgeDays: maxAgeDays,
		Expired:    expiry.Expired,
		Status:     status,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling expire event: %w", err)
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}
