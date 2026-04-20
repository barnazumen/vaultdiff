// Package vault provides utilities for interacting with HashiCorp Vault.
package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TokenRenewResult holds the result of a token renewal operation.
type TokenRenewResult struct {
	ClientToken   string        `json:"client_token"`
	LeaseDuration time.Duration `json:"lease_duration"`
	Renewable     bool          `json:"renewable"`
	Policies      []string      `json:"policies"`
}

// vaultTokenRenewResponse mirrors the Vault API response for token renewal.
type vaultTokenRenewResponse struct {
	Auth struct {
		ClientToken   string   `json:"client_token"`
		LeaseDuration int      `json:"lease_duration"`
		Renewable     bool     `json:"renewable"`
		Policies      []string `json:"policies"`
	} `json:"auth"`
}

// RenewToken attempts to renew the given Vault token, optionally requesting
// a specific increment (in seconds). If increment is 0, Vault uses the
// token's default TTL.
func (c *Client) RenewToken(ctx context.Context, token string, incrementSeconds int) (*TokenRenewResult, error) {
	url := fmt.Sprintf("%s/v1/auth/token/renew-self", c.Address)

	payload := map[string]interface{}{}
	if incrementSeconds > 0 {
		payload["increment"] = fmt.Sprintf("%ds", incrementSeconds)
	}

	body, err := jsonMarshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal renew payload: %w", err)
	}

	req, err := newJSONRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("create renew request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("renew token request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// handled below
	case http.StatusForbidden, http.StatusUnauthorized:
		return nil, fmt.Errorf("renew token: invalid or expired token (HTTP %d)", resp.StatusCode)
	case http.StatusBadRequest:
		return nil, fmt.Errorf("renew token: token is not renewable (HTTP %d)", resp.StatusCode)
	default:
		return nil, fmt.Errorf("renew token: unexpected status %d", resp.StatusCode)
	}

	var vaultResp vaultTokenRenewResponse
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("decode renew response: %w", err)
	}

	return &TokenRenewResult{
		ClientToken:   vaultResp.Auth.ClientToken,
		LeaseDuration: time.Duration(vaultResp.Auth.LeaseDuration) * time.Second,
		Renewable:     vaultResp.Auth.Renewable,
		Policies:      vaultResp.Auth.Policies,
	}, nil
}
