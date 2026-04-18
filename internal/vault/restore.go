package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RestoreFromSnapshot writes all versions from a snapshot back to Vault.
func (c *Client) RestoreFromSnapshot(mountPath, secretPath string, snapshot []SecretSnapshot) error {
	for _, entry := range snapshot {
		if err := c.writeVersion(mountPath, secretPath, entry.Data); err != nil {
			return fmt.Errorf("restore version %d: %w", entry.Version, err)
		}
	}
	return nil
}

func (c *Client) writeVersion(mountPath, secretPath string, data map[string]interface{}) error {
	url := fmt.Sprintf("%s/v1/%s/data/%s", c.address, mountPath, secretPath)

	body := map[string]interface{}{
		"data": data,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := newJSONRequest(http.MethodPost, url, b)
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
