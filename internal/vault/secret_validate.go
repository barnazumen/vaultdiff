package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ValidationResult holds the result of a secret schema validation.
type ValidationResult struct {
	Path       string
	Version    int
	Valid      bool
	Missing    []string
	Extra      []string
	Data       map[string]interface{}
}

// ValidateSecretKeys checks that a secret at the given path and version
// contains all required keys and optionally flags unexpected keys.
func ValidateSecretKeys(addr, token, mount, secretPath string, version int, requiredKeys []string, strictMode bool) (*ValidationResult, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s?version=%d", addr, mount, secretPath, version)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", secretPath)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var envelope struct {
		Data struct {
			Data    map[string]interface{} `json:"data"`
			Metadata struct {
				Version int `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	data := envelope.Data.Data
	requiredSet := make(map[string]struct{}, len(requiredKeys))
	for _, k := range requiredKeys {
		requiredSet[k] = struct{}{}
	}

	var missing, extra []string
	for _, k := range requiredKeys {
		if _, ok := data[k]; !ok {
			missing = append(missing, k)
		}
	}
	if strictMode {
		for k := range data {
			if _, ok := requiredSet[k]; !ok {
				extra = append(extra, k)
			}
		}
	}

	return &ValidationResult{
		Path:    secretPath,
		Version: envelope.Data.Metadata.Version,
		Valid:   len(missing) == 0 && len(extra) == 0,
		Missing: missing,
		Extra:   extra,
		Data:    data,
	}, nil
}
