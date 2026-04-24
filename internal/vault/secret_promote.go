package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PromoteResult holds the result of a secret promotion between mounts.
type PromoteResult struct {
	SourceMount string
	DestMount   string
	Path        string
	Version     int
	Keys        []string
}

// PromoteSecret reads a secret from a source mount and writes it to a
// destination mount, optionally at a different path.
func PromoteSecret(addr, token, srcMount, dstMount, path, dstPath string) (*PromoteResult, error) {
	if dstPath == "" {
		dstPath = path
	}

	// Read from source
	url := fmt.Sprintf("%s/v1/%s/data/%s", addr, srcMount, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("promote: build read request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("promote: read source: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("promote: source secret not found: %s/%s", srcMount, path)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("promote: invalid token or insufficient permissions")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("promote: unexpected status %d reading source", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var srcResp struct {
		Data struct {
			Data     map[string]interface{} `json:"data"`
			Metadata struct {
				Version int `json:"version"`
			} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &srcResp); err != nil {
		return nil, fmt.Errorf("promote: parse source response: %w", err)
	}

	// Write to destination
	payload := map[string]interface{}{"data": srcResp.Data.Data}
	payloadBytes, _ := json.Marshal(payload)

	writeURL := fmt.Sprintf("%s/v1/%s/data/%s", addr, dstMount, dstPath)
	wreq, err := http.NewRequest(http.MethodPost, writeURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, fmt.Errorf("promote: build write request: %w", err)
	}
	wreq.Header.Set("X-Vault-Token", token)
	wreq.Header.Set("Content-Type", "application/json")

	wresp, err := http.DefaultClient.Do(wreq)
	if err != nil {
		return nil, fmt.Errorf("promote: write destination: %w", err)
	}
	defer wresp.Body.Close()

	if wresp.StatusCode != http.StatusOK && wresp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("promote: unexpected status %d writing destination", wresp.StatusCode)
	}

	keys := make([]string, 0, len(srcResp.Data.Data))
	for k := range srcResp.Data.Data {
		keys = append(keys, k)
	}

	return &PromoteResult{
		SourceMount: srcMount,
		DestMount:   dstMount,
		Path:        path,
		Version:     srcResp.Data.Metadata.Version,
		Keys:        keys,
	}, nil
}
