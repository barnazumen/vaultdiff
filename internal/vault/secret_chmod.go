package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SecretPermissions represents the access control permissions for a secret.
type SecretPermissions struct {
	Owner      string   `json:"owner"`
	ReadRoles  []string `json:"read_roles"`
	WriteRoles []string `json:"write_roles"`
	DenyRoles  []string `json:"deny_roles"`
}

// GetSecretPermissions reads the custom-metadata permissions block for a secret path.
func (c *Client) GetSecretPermissions(mount, path string) (*SecretPermissions, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s/%s", mount, path)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result struct {
		Data struct {
			CustomMetadata map[string]string `json:"custom_metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	perms := &SecretPermissions{}
	if v, ok := result.Data.CustomMetadata["owner"]; ok {
		perms.Owner = v
	}
	if v, ok := result.Data.CustomMetadata["read_roles"]; ok && v != "" {
		perms.ReadRoles = splitCSV(v)
	}
	if v, ok := result.Data.CustomMetadata["write_roles"]; ok && v != "" {
		perms.WriteRoles = splitCSV(v)
	}
	if v, ok := result.Data.CustomMetadata["deny_roles"]; ok && v != "" {
		perms.DenyRoles = splitCSV(v)
	}
	return perms, nil
}

// SetSecretPermissions writes permissions into the custom-metadata of a secret.
func (c *Client) SetSecretPermissions(mount, path string, perms SecretPermissions) error {
	cm := map[string]string{
		"owner":       perms.Owner,
		"read_roles":  joinCSV(perms.ReadRoles),
		"write_roles": joinCSV(perms.WriteRoles),
		"deny_roles":  joinCSV(perms.DenyRoles),
	}

	payload, err := json.Marshal(map[string]interface{}{"custom_metadata": cm})
	if err != nil {
		return fmt.Errorf("encode payload: %w", err)
	}

	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.Address, mount, path)
	req, err := http.NewRequest(http.MethodPatch, url, bytesReader(payload))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)
	req.Header.Set("Content-Type", "application/merge-patch+json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("permission denied")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range splitString(s, ',') {
		if t := trimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func splitString(s string, sep rune) []string {
	var parts []string
	start := 0
	for i, r := range s {
		if r == sep {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func joinCSV(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ","
		}
		result += p
	}
	return result
}
