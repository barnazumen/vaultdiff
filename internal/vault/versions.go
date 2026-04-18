package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// VersionMeta holds metadata about a single secret version.
type VersionMeta struct {
	Version      int    `json:"version"`
	CreatedTime  string `json:"created_time"`
	DeletionTime string `json:"deletion_time"`
	Destroyed    bool   `json:"destroyed"`
}

// ListVersions returns metadata for all versions of a KVv2 secret.
func (c *Client) ListVersions(mountPath, secretPath string) ([]VersionMeta, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mountPath, secretPath)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s/%s", mountPath, secretPath)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Versions map[string]struct {
				CreatedTime  string `json:"created_time"`
				DeletionTime string `json:"deletion_time"`
				Destroyed    bool   `json:"destroyed"`
			} `json:"versions"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var versions []VersionMeta
	for k, v := range body.Data.Versions {
		var num int
		fmt.Sscanf(k, "%d", &num)
		versions = append(versions, VersionMeta{
			Version:      num,
			CreatedTime:  v.CreatedTime,
			DeletionTime: v.DeletionTime,
			Destroyed:    v.Destroyed,
		})
	}
	return versions, nil
}
