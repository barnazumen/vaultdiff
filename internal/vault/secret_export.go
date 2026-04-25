package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ExportEntry represents a single exported secret with its metadata.
type ExportEntry struct {
	Path      string                 `json:"path"`
	Version   int                    `json:"version"`
	Data      map[string]interface{} `json:"data"`
	ExportedAt time.Time             `json:"exported_at"`
}

// ExportResult holds all exported entries and summary info.
type ExportResult struct {
	Mount   string        `json:"mount"`
	Entries []ExportEntry `json:"entries"`
	Total   int           `json:"total"`
}

// ExportSecrets reads all listed secrets under a mount and writes them to a JSON file.
func ExportSecrets(addr, token, mount, outputPath string) (*ExportResult, error) {
	client := NewClient(addr, token)

	secrets, err := ListSecrets(client, mount, "")
	if err != nil {
		return nil, fmt.Errorf("listing secrets: %w", err)
	}

	result := &ExportResult{
		Mount:   mount,
		Entries: make([]ExportEntry, 0, len(secrets)),
	}

	for _, path := range secrets {
		data, version, err := readSecretLatest(client, mount, path)
		if err != nil {
			continue
		}
		result.Entries = append(result.Entries, ExportEntry{
			Path:       path,
			Version:    version,
			Data:       data,
			ExportedAt: time.Now().UTC(),
		})
	}

	result.Total = len(result.Entries)

	f, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return nil, fmt.Errorf("encoding export: %w", err)
	}

	return result, nil
}

func readSecretLatest(client *http.Client, mount, path string) (map[string]interface{}, int, error) {
	// Re-use the existing ReadSecretVersion logic with version 0 (latest)
	data, err := ReadSecretVersion(client, mount, path, 0)
	if err != nil {
		return nil, 0, err
	}
	version := 0
	if v, ok := data["version"].(float64); ok {
		version = int(v)
	}
	delete(data, "version")
	return data, version, nil
}
