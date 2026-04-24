package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SecretLockInfo holds lock metadata for a secret path.
type SecretLockInfo struct {
	Path      string    `json:"path"`
	LockedBy  string    `json:"locked_by"`
	LockedAt  time.Time `json:"locked_at"`
	Reason    string    `json:"reason"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// LockSecret writes a lock marker into the secret's metadata custom_metadata.
func (c *Client) LockSecret(path, mount, token, reason string, ttlSeconds int) (*SecretLockInfo, error) {
	now := time.Now().UTC()
	info := SecretLockInfo{
		Path:     path,
		LockedBy: token,
		LockedAt: now,
		Reason:   reason,
	}
	if ttlSeconds > 0 {
		info.ExpiresAt = now.Add(time.Duration(ttlSeconds) * time.Second)
	}

	body := map[string]interface{}{
		"custom_metadata": map[string]string{
			"locked":     "true",
			"locked_by":  token,
			"locked_at":  now.Format(time.RFC3339),
			"lock_reason": reason,
		},
	}
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, path)
	if err := c.postJSON(url, token, body); err != nil {
		return nil, fmt.Errorf("lock secret: %w", err)
	}
	return &info, nil
}

// UnlockSecret removes the lock marker from the secret's metadata.
func (c *Client) UnlockSecret(path, mount, token string) error {
	body := map[string]interface{}{
		"custom_metadata": map[string]string{
			"locked":     "",
			"locked_by":  "",
			"locked_at":  "",
			"lock_reason": "",
		},
	}
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, path)
	if err := c.postJSON(url, token, body); err != nil {
		return fmt.Errorf("unlock secret: %w", err)
	}
	return nil
}

// IsLocked checks whether a secret path currently has a lock marker.
func (c *Client) IsLocked(path, mount, token string) (bool, *SecretLockInfo, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, nil, err
	}
	req.Header.Set("X-Vault-Token", token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return false, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var result struct {
		Data struct {
			CustomMetadata map[string]string `json:"custom_metadata"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil, err
	}
	cm := result.Data.CustomMetadata
	if cm["locked"] != "true" {
		return false, nil, nil
	}
	lockedAt, _ := time.Parse(time.RFC3339, cm["locked_at"])
	info := &SecretLockInfo{
		Path:     path,
		LockedBy: cm["locked_by"],
		LockedAt: lockedAt,
		Reason:   cm["lock_reason"],
	}
	return true, info, nil
}
