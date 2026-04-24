package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SecretExpiry holds expiration metadata for a KV secret version.
type SecretExpiry struct {
	Path        string
	Version     int
	CreatedTime time.Time
	DaysOld     int
	Expired     bool
}

// CheckSecretExpiry reads a secret's metadata and computes expiry info.
func (c *Client) CheckSecretExpiry(mount, secretPath string, maxAgeDays int) (*SecretExpiry, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, secretPath)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
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

	var result struct {
		Data struct {
			CurrentVersion int `json:"current_version"`
			Versions       map[string]struct {
				CreatedTime string `json:"created_time"`
			} `json:"versions"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	ver := result.Data.CurrentVersion
	key := fmt.Sprintf("%d", ver)
	versionData, ok := result.Data.Versions[key]
	if !ok {
		return nil, fmt.Errorf("version %d metadata not found", ver)
	}

	created, err := time.Parse(time.RFC3339Nano, versionData.CreatedTime)
	if err != nil {
		return nil, fmt.Errorf("parsing created_time: %w", err)
	}

	daysOld := int(time.Since(created).Hours() / 24)
	return &SecretExpiry{
		Path:        secretPath,
		Version:     ver,
		CreatedTime: created,
		DaysOld:     daysOld,
		Expired:     daysOld > maxAgeDays,
	}, nil
}
