package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AuditTrailEntry represents a single audit trail record for a secret.
type AuditTrailEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Version   int       `json:"version"`
	Actor     string    `json:"actor"`
	Path      string    `json:"path"`
}

// AuditTrail holds all audit trail entries for a secret.
type AuditTrail struct {
	Path    string            `json:"path"`
	Entries []AuditTrailEntry `json:"entries"`
}

// ReadAuditTrail fetches the audit trail for a given secret path from Vault metadata.
func (c *Client) ReadAuditTrail(mount, secretPath string) (*AuditTrail, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, secretPath)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", secretPath)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: invalid token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var raw struct {
		Data struct {
			Versions map[string]struct {
				CreatedTime  time.Time `json:"created_time"`
				DeletionTime string    `json:"deletion_time"`
				Destroyed    bool      `json:"destroyed"`
			} `json:"versions"`
			CreatedTime time.Time `json:"created_time"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	trail := &AuditTrail{Path: secretPath}
	for versionKey, vdata := range raw.Data.Versions {
		var vNum int
		fmt.Sscanf(versionKey, "%d", &vNum)
		action := "created"
		if vdata.Destroyed {
			action = "destroyed"
		} else if vdata.DeletionTime != "" {
			action = "deleted"
		}
		trail.Entries = append(trail.Entries, AuditTrailEntry{
			Timestamp: vdata.CreatedTime,
			Action:    action,
			Version:   vNum,
			Actor:     "vault",
			Path:      secretPath,
		})
	}
	return trail, nil
}
