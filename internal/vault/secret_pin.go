package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PinResult holds the outcome of a pin or unpin operation.
type PinResult struct {
	Path    string
	Version int
	Pinned  bool
}

// PinSecret marks a specific version of a KV v2 secret as pinned by writing
// a custom-metadata flag "pinned_version" on the secret's metadata.
func PinSecret(addr, token, mount, secretPath string, version int) (*PinResult, error) {
	client := NewClient(addr, token)

	body := map[string]interface{}{
		"custom_metadata": map[string]string{
			"pinned_version": fmt.Sprintf("%d", version),
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("pin: marshal body: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s/metadata/%s", addr, mount, secretPath)
	req, err := http.NewRequest(http.MethodPost, url, bytesReader(b))
	if err != nil {
		return nil, fmt.Errorf("pin: create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pin: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("pin: invalid token")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("pin: secret not found: %s", secretPath)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pin: unexpected status %d: %s", resp.StatusCode, string(raw))
	}

	return &PinResult{Path: secretPath, Version: version, Pinned: true}, nil
}

// UnpinSecret removes the pinned_version custom-metadata flag from a secret.
func UnpinSecret(addr, token, mount, secretPath string) (*PinResult, error) {
	client := NewClient(addr, token)

	body := map[string]interface{}{
		"custom_metadata": map[string]string{
			"pinned_version": "",
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("unpin: marshal body: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s/metadata/%s", addr, mount, secretPath)
	req, err := http.NewRequest(http.MethodPost, url, bytesReader(b))
	if err != nil {
		return nil, fmt.Errorf("unpin: create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unpin: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unpin: invalid token")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("unpin: secret not found: %s", secretPath)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unpin: unexpected status %d: %s", resp.StatusCode, string(raw))
	}

	return &PinResult{Path: secretPath, Version: 0, Pinned: false}, nil
}
