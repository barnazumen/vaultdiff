package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Annotation represents a key-value annotation attached to a secret path.
type Annotation struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AnnotationResult holds the current annotations for a secret.
type AnnotationResult struct {
	Path        string            `json:"path"`
	Annotations map[string]string `json:"annotations"`
}

// SetAnnotation writes a single annotation (stored as custom_metadata) to the
// KV v2 metadata endpoint for the given secret path.
func (c *Client) SetAnnotation(mount, path, key, value string) error {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.addr, mount, path)

	body := map[string]interface{}{
		"custom_metadata": map[string]string{
			key: value,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal annotation: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(string(b)))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/merge-patch+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("set annotation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("permission denied")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(data))
	}
	return nil
}

// GetAnnotations reads the custom_metadata for the given secret path and
// returns it as an AnnotationResult.
func (c *Client) GetAnnotations(mount, path string) (*AnnotationResult, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.addr, mount, path)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get annotations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied")
	}

	var out struct {
		Data struct {
			CustomMetadata map[string]string `json:"custom_metadata"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	anns := out.Data.CustomMetadata
	if anns == nil {
		anns = map[string]string{}
	}
	return &AnnotationResult{Path: path, Annotations: anns}, nil
}
