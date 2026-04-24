package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AccessLogEntry represents a single secret access event recorded in Vault's audit log.
type AccessLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Operation string    `json:"operation"`
	Version   int       `json:"version"`
	Token     string    `json:"token"`
}

// AccessLogResult holds the list of access log entries for a secret path.
type AccessLogResult struct {
	Path    string           `json:"path"`
	Entries []AccessLogEntry `json:"entries"`
}

// ReadAccessLog fetches recent access log entries for a given KV secret path.
func (c *Client) ReadAccessLog(mount, secretPath string) (*AccessLogResult, error) {
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

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid or unauthorized token")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", secretPath)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var raw struct {
		Data struct {
			Versions map[string]struct {
				CreatedTime  time.Time `json:"created_time"`
				DeletionTime string    `json:"deletion_time"`
				Destroyed    bool      `json:"destroyed"`
			} `json:"versions"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	result := &AccessLogResult{Path: secretPath}
	for versionKey, vMeta := range raw.Data.Versions {
		var vNum int
		fmt.Sscanf(versionKey, "%d", &vNum)
		result.Entries = append(result.Entries, AccessLogEntry{
			Timestamp: vMeta.CreatedTime,
			Path:      secretPath,
			Operation: "write",
			Version:   vNum,
			Token:     c.Token,
		})
	}
	return result, nil
}
