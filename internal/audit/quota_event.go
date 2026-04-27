package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// QuotaEvent records an audit entry when quota information is read for a path.
type QuotaEvent struct {
	Timestamp     string  `json:"timestamp"`
	Event         string  `json:"event"`
	QuotaName     string  `json:"quota_name"`
	Path          string  `json:"path"`
	Type          string  `json:"type"`
	MaxLeases     int     `json:"max_leases"`
	CurrentLeases int     `json:"current_leases"`
	Rate          float64 `json:"rate"`
	Burst         int     `json:"burst"`
}

// LogQuotaEvent writes a quota audit event to the provided writer.
func LogQuotaEvent(w io.Writer, quotaName, path, qtype string, maxLeases, currentLeases int, rate float64, burst int) error {
	event := QuotaEvent{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Event:         "quota_read",
		QuotaName:     quotaName,
		Path:          path,
		Type:          qtype,
		MaxLeases:     maxLeases,
		CurrentLeases: currentLeases,
		Rate:          rate,
		Burst:         burst,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling quota event: %w", err)
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}
