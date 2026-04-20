package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TokenLookup holds metadata about a Vault token.
type TokenLookup struct {
	Accessor   string   `json:"accessor"`
	Policies   []string `json:"policies"`
	TTL        int      `json:"ttl"`
	ExpireTime string   `json:"expire_time"`
	DisplayName string  `json:"display_name"`
}

// LookupToken calls the Vault token self-lookup endpoint and returns metadata.
func (c *Client) LookupToken() (*TokenLookup, error) {
	url := fmt.Sprintf("%s/v1/auth/token/lookup-self", c.Address)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	var result struct {
		Data TokenLookup `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &result.Data, nil
}
