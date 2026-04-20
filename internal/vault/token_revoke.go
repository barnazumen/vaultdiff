package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// RevokeToken revokes a specific token via the Vault API.
// If selfRevoke is true, it calls revoke-self using the client token.
func (c *Client) RevokeToken(token string, selfRevoke bool) error {
	var (
		path string
		body io.Reader
	)

	if selfRevoke {
		path = "/v1/auth/token/revoke-self"
	} else {
		path = "/v1/auth/token/revoke"
		payload, err := json.Marshal(map[string]string{"token": token})
		if err != nil {
			return fmt.Errorf("encoding payload: %w", err)
		}
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(http.MethodPost, c.Address+path, body)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)
	if !selfRevoke {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("revoking token: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusOK:
		return nil
	case http.StatusForbidden:
		return fmt.Errorf("permission denied: invalid or insufficient token")
	default:
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}
