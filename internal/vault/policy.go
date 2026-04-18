package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// PolicyInfo holds the rules associated with a Vault policy.
type PolicyInfo struct {
	Name  string `json:"name"`
	Rules string `json:"rules"`
}

// policyResponse mirrors the Vault API response for a policy read.
type policyResponse struct {
	Data PolicyInfo `json:"data"`
}

// ReadPolicy fetches the policy rules for the given policy name from Vault.
func (c *Client) ReadPolicy(policyName string) (*PolicyInfo, error) {
	url := fmt.Sprintf("%s/v1/sys/policies/acl/%s", c.address, policyName)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusNotFound:
		return nil, fmt.Errorf("policy %q not found", policyName)
	case http.StatusForbidden:
		return nil, fmt.Errorf("permission denied: invalid token")
	default:
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result policyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result.Data, nil
}
