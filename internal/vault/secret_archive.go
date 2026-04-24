package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ArchiveResult holds the result of archiving a secret version.
type ArchiveResult struct {
	Path    string `json:"path"`
	Version int    `json:"version"`
	Archived bool   `json:"archived"`
}

// ArchiveSecretVersion soft-deletes (archives) a specific version of a KV v2 secret.
func (c *Client) ArchiveSecretVersion(mount, secretPath string, version int) (*ArchiveResult, error) {
	url := fmt.Sprintf("%s/v1/%s/delete/%s", c.Address, mount, secretPath)

	body := map[string]interface{}{
		"versions": []int{version},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("archive: marshal body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytesReader(data))
	if err != nil {
		return nil, fmt.Errorf("archive: create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("archive: request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusOK:
		return &ArchiveResult{
			Path:     secretPath,
			Version:  version,
			Archived: true,
		}, nil
	case http.StatusForbidden:
		return nil, fmt.Errorf("archive: permission denied (invalid token)")
	case http.StatusNotFound:
		return nil, fmt.Errorf("archive: secret not found: %s", secretPath)
	default:
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("archive: unexpected status %d: %s", resp.StatusCode, string(b))
	}
}
