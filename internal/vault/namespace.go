package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// NamespaceInfo holds metadata about a Vault namespace.
type NamespaceInfo struct {
	Path        string            `json:"path"`
	ID          string            `json:"id"`
	CustomMetadata map[string]string `json:"custom_metadata"`
}

// ListNamespaces returns all child namespaces under the given prefix.
// Pass an empty prefix to list root-level namespaces.
func (c *Client) ListNamespaces(prefix string) ([]NamespaceInfo, error) {
	path := "/v1/sys/namespaces"
	if prefix != "" {
		path = fmt.Sprintf("/v1/sys/namespaces/%s", prefix)
	}

	req, err := http.NewRequest(http.MethodGet, c.address+path, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("X-Vault-Request", "true")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid or insufficient token")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("namespace path not found: %s", prefix)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var body struct {
		Data map[string]struct {
			ID             string            `json:"id"`
			CustomMetadata map[string]string `json:"custom_metadata"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var namespaces []NamespaceInfo
	for path, info := range body.Data {
		namespaces = append(namespaces, NamespaceInfo{
			Path:           path,
			ID:             info.ID,
			CustomMetadata: info.CustomMetadata,
		})
	}
	return namespaces, nil
}
