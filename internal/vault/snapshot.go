package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// SecretSnapshot holds a point-in-time export of a secret's versions.
type SecretSnapshot struct {
	Path      string                 `json:"path"`
	ExportedAt time.Time             `json:"exported_at"`
	Versions  map[int]SecretVersion  `json:"versions"`
}

// SecretVersion holds the data for a single secret version.
type SecretVersion struct {
	Version  int                    `json:"version"`
	Data     map[string]interface{} `json:"data"`
	CreatedAt time.Time             `json:"created_at"`
}

// ExportSnapshot reads all available versions for a path and writes a JSON snapshot.
func (c *Client) ExportSnapshot(path string, w io.Writer) error {
	versions, err := c.ListVersions(path)
	if err != nil {
		return fmt.Errorf("list versions: %w", err)
	}

	snap := SecretSnapshot{
		Path:       path,
		ExportedAt: time.Now().UTC(),
		Versions:   make(map[int]SecretVersion),
	}

	for _, v := range versions {
		data, err := c.ReadSecretVersion(path, v)
		if err != nil {
			return fmt.Errorf("read version %d: %w", v, err)
		}
		snap.Versions[v] = SecretVersion{
			Version: v,
			Data:    data,
		}
	}

	return json.NewEncoder(w).Encode(snap)
}

// LoadSnapshot reads a SecretSnapshot from a JSON file on disk.
func LoadSnapshot(filePath string) (*SecretSnapshot, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open snapshot: %w", err)
	}
	defer f.Close()

	var snap SecretSnapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	return &snap, nil
}
