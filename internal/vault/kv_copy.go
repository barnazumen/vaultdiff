package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// KVCopyOptions controls the behaviour of CopySecret.
type KVCopyOptions struct {
	SourceMount string
	DestMount   string
	SourcePath  string
	DestPath    string
	Version     int // 0 means latest
}

// CopySecret reads a KV-v2 secret from a source path and writes it to a
// destination path, optionally pinning a specific source version.
func (c *Client) CopySecret(opts KVCopyOptions) error {
	srcMount := opts.SourceMount
	if srcMount == "" {
		srcMount = "secret"
	}
	dstMount := opts.DestMount
	if dstMount == "" {
		dstMount = srcMount
	}

	// Read source
	url := fmt.Sprintf("%s/v1/%s/data/%s", c.addr, srcMount, opts.SourcePath)
	if opts.Version > 0 {
		url += fmt.Sprintf("?version=%d", opts.Version)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("kv_copy: build read request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("kv_copy: read request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("kv_copy: source secret not found: %s", opts.SourcePath)
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("kv_copy: permission denied reading %s", opts.SourcePath)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kv_copy: unexpected status %d reading source", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("kv_copy: read body: %w", err)
	}

	var envelope struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("kv_copy: unmarshal source: %w", err)
	}

	// Write destination
	return c.writeKVData(dstMount, opts.DestPath, envelope.Data.Data)
}

func (c *Client) writeKVData(mount, path string, data map[string]interface{}) error {
	payload, err := json.Marshal(map[string]interface{}{"data": data})
	if err != nil {
		return fmt.Errorf("kv_copy: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s/data/%s", c.addr, mount, path)
	req, err := http.NewRequest(http.MethodPost, url, bytesReader(payload))
	if err != nil {
		return fmt.Errorf("kv_copy: build write request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("kv_copy: write request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("kv_copy: permission denied writing %s", path)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("kv_copy: unexpected status %d writing destination", resp.StatusCode)
	}
	return nil
}
