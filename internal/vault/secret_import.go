package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ImportEntry represents a single secret to import.
type ImportEntry struct {
	Path string            `json:"path"`
	Data map[string]string `json:"data"`
}

// ImportResult holds the outcome of a single secret import.
type ImportResult struct {
	Path    string
	Success bool
	Error   string
}

// ImportSecretsFromFile reads a JSON file containing a list of ImportEntry
// values and writes each secret to Vault KV v2.
func ImportSecretsFromFile(addr, token, mount, filePath string) ([]ImportResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open import file: %w", err)
	}
	defer f.Close()

	var entries []ImportEntry
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return nil, fmt.Errorf("decode import file: %w", err)
	}

	var results []ImportResult
	for _, e := range entries {
		res := ImportResult{Path: e.Path}
		if err := writeImportEntry(addr, token, mount, e); err != nil {
			res.Success = false
			res.Error = err.Error()
		} else {
			res.Success = true
		}
		results = append(results, res)
	}
	return results, nil
}

func writeImportEntry(addr, token, mount string, e ImportEntry) error {
	payload := map[string]interface{}{"data": e.Data}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/v1/%s/data/%s", addr, mount, e.Path)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid token")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault returned %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
