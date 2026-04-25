package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TouchResult holds the result of a secret touch (re-write) operation.
type TouchResult struct {
	Path    string
	Version int
	Mount   string
}

// TouchSecret re-writes the latest version of a secret at the given path,
// effectively creating a new version with identical data. This is useful for
// resetting TTLs or triggering watch-based workflows.
func TouchSecret(addr, token, mount, secretPath string) (*TouchResult, error) {
	client := NewClient(addr, token)

	// Read the current latest data.
	readURL := fmt.Sprintf("%s/v1/%s/data/%s", addr, mount, secretPath)
	req, err := http.NewRequest(http.MethodGet, readURL, nil)
	if err != nil {
		return nil, fmt.Errorf("touch: build read request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("touch: read request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("touch: secret not found: %s", secretPath)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("touch: permission denied")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("touch: unexpected status %d", resp.StatusCode)
	}

	var readResp struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&readResp); err != nil {
		return nil, fmt.Errorf("touch: decode read response: %w", err)
	}

	// Write the same data back to create a new version.
	writeURL := fmt.Sprintf("%s/v1/%s/data/%s", addr, mount, secretPath)
	payload := map[string]interface{}{"data": readResp.Data.Data}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("touch: marshal write payload: %w", err)
	}

	wreq, err := http.NewRequest(http.MethodPost, writeURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("touch: build write request: %w", err)
	}
	wreq.Header.Set("X-Vault-Token", token)
	wreq.Header.Set("Content-Type", "application/json")

	wresp, err := client.HTTPClient.Do(wreq)
	if err != nil {
		return nil, fmt.Errorf("touch: write request: %w", err)
	}
	defer wresp.Body.Close()

	if wresp.StatusCode != http.StatusOK && wresp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(wresp.Body)
		return nil, fmt.Errorf("touch: write failed (%d): %s", wresp.StatusCode, string(body))
	}

	var writeResp struct {
		Data struct {
			Version int `json:"version"`
		} `json:"data"`
	}
	_ = json.NewDecoder(wresp.Body).Decode(&writeResp)

	return &TouchResult{
		Path:    secretPath,
		Version: writeResp.Data.Version,
		Mount:   mount,
	}, nil
}
