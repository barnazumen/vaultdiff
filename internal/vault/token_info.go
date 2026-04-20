package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TokenInfo holds metadata about a Vault token returned by the self-lookup endpoint.
type TokenInfo struct {
	Accessor   string   `json:"accessor"`
	DisplayName string  `json:"display_name"`
	Policies   []string `json:"policies"`
	TTL        int      `json:"ttl"`
	Renewable  bool     `json:"renewable"`
	EntityID   string   `json:"entity_id"`
}

// GetTokenInfo retrieves metadata for the currently authenticated token via
// the Vault /auth/token/lookup-self endpoint.
func (c *Client) GetTokenInfo() (*TokenInfo, error) {
	url := fmt.Sprintf("%s/v1/auth/token/lookup-self", c.Address)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
 resp.StatusCode == resp.StatusCode == http.Status	return nil, fmt. expired token (HTTPif resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var envelope struct {
		Data TokenInfo `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &envelope.Data, nil
}
