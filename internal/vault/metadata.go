package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VersionMetadata holds metadata for a single secret version.
type VersionMetadata struct {
	Version      int       `json:"version"`
	CreatedTime  time.Time `json:"created_time"`
	DeletionTime time.Time `json:"deletion_time"`
	Destroyed    bool      `json:"destroyed"`
}

// SecretMetadata holds full metadata for a KV v2 secret path.
type SecretMetadata struct {
	Path            string                     `json:"path"`
	CurrentVersion  int                        `json:"current_version"`
	OldestVersion   int                        `json:"oldest_version"`
	CreatedTime     time.Time                  `json:"created_time"`
	Versions        map[string]VersionMetadata `json:"versions"`
}

// ReadSecretMetadata fetches metadata for a KV v2 secret path.
func (c *Client) ReadSecretMetadata(mount, secretPath string) (*SecretMetadata, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.address, mount, secretPath)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", secretPath)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var envelope struct {
		Data struct {
			CurrentVersion int                        `json:"current_version"`
			OldestVersion  int                        `json:"oldest_version"`
			CreatedTime    time.Time                  `json:"created_time"`
			Versions       map[string]VersionMetadata `json:"versions"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &SecretMetadata{
		Path:           secretPath,
		CurrentVersion: envelope.Data.CurrentVersion,
		OldestVersion:  envelope.Data.OldestVersion,
		CreatedTime:    envelope.Data.CreatedTime,
		Versions:       envelope.Data.Versions,
	}, nil
}
