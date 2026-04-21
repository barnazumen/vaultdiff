package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TokenTTLInfo holds TTL and expiration details for a Vault token.
type TokenTTLInfo struct {
	TTL            int       `json:"ttl"`
	CreationTTL    int       `json:"creation_ttl"`
	ExpireTime     time.Time `json:"expire_time"`
	ExplicitMaxTTL int       `json:"explicit_max_ttl"`
	Period         int       `json:"period"`
}

// GetTokenTTL fetches TTL information for the current token.
func (c *Client) GetTokenTTL() (*TokenTTLInfo, error) {
	req, err := http.NewRequest(http.MethodGet, c.addr+"/v1/auth/token/lookup-self", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var result struct {
		Data struct {
			TTL            int    `json:"ttl"`
			CreationTTL    int    `json:"creation_ttl"`
			ExpireTime     string `json:"expire_time"`
			ExplicitMaxTTL int    `json:"explicit_max_ttl"`
			Period         int    `json:"period"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	info := &TokenTTLInfo{
		TTL:            result.Data.TTL,
		CreationTTL:    result.Data.CreationTTL,
		ExplicitMaxTTL: result.Data.ExplicitMaxTTL,
		Period:         result.Data.Period,
	}

	if result.Data.ExpireTime != "" {
		t, err := time.Parse(time.RFC3339, result.Data.ExpireTime)
		if err == nil {
			info.ExpireTime = t
		}
	}

	return info, nil
}
