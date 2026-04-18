package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// PolicyAuditEntry records a policy comparison event.
type PolicyAuditEntry struct {
	Timestamp  string   `json:"timestamp"`
	PolicyA    string   `json:"policy_a"`
	PolicyB    string   `json:"policy_b"`
	Added      []string `json:"added"`
	Removed    []string `json:"removed"`
	Unchanged  []string `json:"unchanged"`
	HasDiff    bool     `json:"has_diff"`
}

// LogPolicyAudit writes a policy diff audit entry as JSON to the given file.
func LogPolicyAudit(path, policyA, policyB string, diff []PolicyLineDiff) error {
	entry := PolicyAuditEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		PolicyA:   policyA,
		PolicyB:   policyB,
	}

	for _, d := range diff {
		switch d.Type {
		case "added":
			entry.Added = append(entry.Added, d.Line)
		case "removed":
			entry.Removed = append(entry.Removed, d.Line)
		case "unchanged":
			entry.Unchanged = append(entry.Unchanged, d.Line)
		}
	}

	entry.HasDiff = len(entry.Added) > 0 || len(entry.Removed) > 0

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(entry)
}
