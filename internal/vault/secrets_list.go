package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SecretList holds the keys returned from a KV v2 list operation.
type SecretList struct {
	Keys []string
}

// ListSecrets returns all secret keys under the given mount and path prefix.
func (c *Client) ListSecrets(mount, path string) (*SecretList, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, path)

	req, err := http.NewRequest("LIST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building list request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing list request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("path not found: %s/%s", mount, path)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: invalid token or policy")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d listing secrets", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding list response: %w", err)
	}

	return &SecretList{Keys: body.Data.Keys}, nil
}
