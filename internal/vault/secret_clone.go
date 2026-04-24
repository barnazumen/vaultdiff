package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CloneSecretOptions configures a clone operation.
type CloneSecretOptions struct {
	SourcePath  string
	DestPath    string
	Mount       string
	Version     int
	Overwrite   bool
}

// CloneResult holds the result of a clone operation.
type CloneResult struct {
	SourcePath string
	DestPath   string
	Version    int
	Keys       []string
}

// CloneSecret reads a specific version of a secret and writes it to a new path.
func CloneSecret(addr, token string, opts CloneSecretOptions) (*CloneResult, error) {
	mount := opts.Mount
	if mount == "" {
		mount = "secret"
	}

	// Read source secret at the given version
	url := fmt.Sprintf("%s/v1/%s/data/%s", addr, mount, opts.SourcePath)
	if opts.Version > 0 {
		url += fmt.Sprintf("?version=%d", opts.Version)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("clone: build read request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("clone: read source: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("clone: source secret not found: %s", opts.SourcePath)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("clone: permission denied reading source")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clone: unexpected status reading source: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("clone: read body: %w", err)
	}

	var envelope struct {
		Data struct {
			Data    map[string]interface{} `json:"data"`
			Metadata struct {
				Version int `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("clone: parse source response: %w", err)
	}

	// Write to destination
	writeURL := fmt.Sprintf("%s/v1/%s/data/%s", addr, mount, opts.DestPath)
	payload := map[string]interface{}{"data": envelope.Data.Data}
	writeBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("clone: marshal write payload: %w", err)
	}

	method := http.MethodPost
	if opts.Overwrite {
		method = http.MethodPut
	}

	wreq, err := http.NewRequest(method, writeURL, bytesReader(writeBody))
	if err != nil {
		return nil, fmt.Errorf("clone: build write request: %w", err)
	}
	wreq.Header.Set("X-Vault-Token", token)
	wreq.Header.Set("Content-Type", "application/json")

	wresp, err := client.Do(wreq)
	if err != nil {
		return nil, fmt.Errorf("clone: write dest: %w", err)
	}
	defer wresp.Body.Close()

	if wresp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("clone: permission denied writing destination")
	}
	if wresp.StatusCode != http.StatusOK && wresp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("clone: unexpected status writing dest: %d", wresp.StatusCode)
	}

	keys := make([]string, 0, len(envelope.Data.Data))
	for k := range envelope.Data.Data {
		keys = append(keys, k)
	}

	return &CloneResult{
		SourcePath: opts.SourcePath,
		DestPath:   opts.DestPath,
		Version:    envelope.Data.Metadata.Version,
		Keys:       keys,
	}, nil
}
