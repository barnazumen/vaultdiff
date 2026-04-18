package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LeaseInfo holds metadata about a Vault secret lease.
type LeaseInfo struct {
	LeaseID       string        `json:"lease_id"`
	Renewable     bool          `json:"renewable"`
	LeaseDuration time.Duration `json:"lease_duration"`
}

type leaseResponse struct {
	LeaseID       string `json:"lease_id"`
	Renewable     bool   `json:"renewable"`
	LeaseDuration int    `json:"lease_duration"`
}

// ReadLeaseInfo fetches lease metadata for a dynamic secret path.
func (c *Client) ReadLeaseInfo(path string) (*LeaseInfo, error) {
	url := fmt.Sprintf("%s/v1/%s", c.address, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var lr leaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &LeaseInfo{
		LeaseID:       lr.LeaseID,
		Renewable:     lr.Renewable,
		LeaseDuration: time.Duration(lr.LeaseDuration) * time.Second,
	}, nil
}
