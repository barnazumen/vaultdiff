package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// write sends a POST/PUT request with a JSON body to the given Vault path.
func (c *Client) write(ctx context.Context, path string, body map[string]any) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s", c.address, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("create write request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	return c.http.Do(req)
}

// delete sends a DELETE request to the given Vault path.
func (c *Client) delete(ctx context.Context, path string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/%s", c.address, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create delete request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	return c.http.Do(req)
}
