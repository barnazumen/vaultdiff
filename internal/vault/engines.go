package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MountInfo represents a single secrets engine mount.
type MountInfo struct {
	Path        string
	Type        string
	Description string
	Options     map[string]string
}

// ListSecretEngines returns all mounted secrets engines from Vault.
func (c *Client) ListSecretEngines() ([]MountInfo, error) {
	req, err := http.NewRequest(http.MethodGet, c.addr+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
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

	var raw map[string]struct {
		Type        string            `json:"type"`
		Description string            `json:"description"`
		Options     map[string]string `json:"options"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var mounts []MountInfo
	for path, info := range raw {
		mounts = append(mounts, MountInfo{
			Path:        path,
			Type:        info.Type,
			Description: info.Description,
			Options:     info.Options,
		})
	}
	return mounts, nil
}
