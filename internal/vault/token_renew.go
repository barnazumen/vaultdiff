package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// TokenRenewResult holds the result of a token renewal operation.
type TokenRenewResult struct {
	ClientToken   string
	LeaseDuration int
	Renewable     bool
}

// RenewToken attempts to renew the given Vault token and returns the updated lease info.
func (c *Client) RenewToken(token string) (*TokenRenewResult, error) {
	url := fmt.Sprintf("%s/v1/auth/token/renew-self", c.Address)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building renew request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("renew request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Auth struct {
			ClientToken   string `json:"client_token"`
			LeaseDuration int    `json:"lease_duration"`
			Renewable     bool   `json:"renewable"`
		} `json:"auth"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding renew response: %w", err)
	}

	return &TokenRenewResult{
		ClientToken:   result.Auth.ClientToken,
		LeaseDuration: result.Auth.LeaseDuration,
		Renewable:     result.Auth.Renewable,
	}, nil
}
