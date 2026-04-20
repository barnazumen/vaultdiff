package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TokenCapabilities holds the list of capabilities a token has on a given path.
type TokenCapabilities struct {
	Path         string   `json:"path"`
	Capabilities []string `json:"capabilities"`
}

// GetTokenCapabilities queries Vault for the capabilities of the current token
// on the given secret path.
func (c *Client) GetTokenCapabilities(path string) (*TokenCapabilities, error) {
	url := fmt.Sprintf("%s/v1/sys/capabilities-self", c.Address)

	payload := fmt.Sprintf(`{"paths":["%s"]}`, path)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting capabilities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("invalid or unauthorized token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result struct {
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &TokenCapabilities{
		Path:         path,
		Capabilities: result.Capabilities,
	}, nil
}
