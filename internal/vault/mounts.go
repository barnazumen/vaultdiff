package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MountInfo holds basic information about a Vault secret mount.
type MountInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Accessor    string `json:"accessor"`
}

// ListMounts returns all secret mounts from Vault.
func (c *Client) ListMounts() (map[string]MountInfo, error) {
	req, err := http.NewRequest("GET", c.Address+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("invalid token or permission denied")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]MountInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse mounts response: %w", err)
	}
	return result, nil
}
