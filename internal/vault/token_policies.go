package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TokenPolicies holds the policies attached to a token.
type TokenPolicies struct {
	IdentityPolicies []string `json:"identity_policies"`
	Policies         []string `json:"policies"`
	TokenPolicies    []string `json:"token_policies"`
}

type tokenPoliciesResponse struct {
	Data struct {
		IdentityPolicies []string `json:"identity_policies"`
		Policies         []string `json:"policies"`
		TokenPolicies    []string `json:"token_policies"`
	} `json:"data"`
}

// GetTokenPolicies retrieves all policies associated with the given token.
func (c *Client) GetTokenPolicies(token string) (*TokenPolicies, error) {
	url := fmt.Sprintf("%s/v1/auth/token/lookup", c.Address)

	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("invalid or unauthorized token")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("token not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result tokenPoliciesResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &TokenPolicies{
		IdentityPolicies: result.Data.IdentityPolicies,
		Policies:         result.Data.Policies,
		TokenPolicies:    result.Data.TokenPolicies,
	}, nil
}
