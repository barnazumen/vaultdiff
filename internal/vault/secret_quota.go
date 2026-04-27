package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// QuotaInfo holds rate limit and lease count quota details for a secret path.
type QuotaInfo struct {
	Path           string  `json:"path"`
	Type           string  `json:"type"`
	MaxLeases      int     `json:"max_leases"`
	CurrentLeases  int     `json:"current_leases"`
	Rate           float64 `json:"rate"`
	Burst          int     `json:"burst"`
	IntervalSeconds int    `json:"interval_seconds"`
}

// ReadSecretQuota fetches quota information for a given path from Vault.
func ReadSecretQuota(vaultAddr, token, quotaName string) (*QuotaInfo, error) {
	url := fmt.Sprintf("%s/v1/sys/quotas/rate-limit/%s", vaultAddr, quotaName)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("quota %q not found", quotaName)
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("permission denied: invalid token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var envelope struct {
		Data QuotaInfo `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &envelope.Data, nil
}
