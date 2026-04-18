package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SecretTags represents custom metadata (tags) for a KV v2 secret.
type SecretTags map[string]string

// ReadSecretTags fetches the custom_metadata field from a KV v2 secret's metadata.
func (c *Client) ReadSecretTags(mount, path string) (SecretTags, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, path)

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
		return nil, fmt.Errorf("secret not found: %s/%s", mount, path)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: invalid token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			CustomMetadata SecretTags `json:"custom_metadata"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if result.Data.CustomMetadata == nil {
		return SecretTags{}, nil
	}
	return result.Data.CustomMetadata, nil
}
