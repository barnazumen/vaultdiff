package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SecretSearchResult holds a matched secret path and its matching keys.
type SecretSearchResult struct {
	Path        string
	MatchingKeys []string
}

// SearchSecrets lists all secrets under the given mount/prefix and returns
// those whose keys or values contain the query string.
func SearchSecrets(addr, token, mount, prefix, query string) ([]SecretSearchResult, error) {
	keys, err := listSecretsRecursive(addr, token, mount, prefix)
	if err != nil {
		return nil, err
	}

	var results []SecretSearchResult
	for _, key := range keys {
		data, err := readSecretData(addr, token, mount, key)
		if err != nil {
			continue
		}
		var matched []string
		for k, v := range data {
			if strings.Contains(k, query) || strings.Contains(fmt.Sprintf("%v", v), query) {
				matched = append(matched, k)
			}
		}
		if len(matched) > 0 {
			results = append(results, SecretSearchResult{Path: key, MatchingKeys: matched})
		}
	}
	return results, nil
}

func listSecretsRecursive(addr, token, mount, prefix string) ([]string, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s?list=true", addr, mount, prefix)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list secrets: unexpected status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var all []string
	for _, k := range result.Data.Keys {
		full := prefix + k
		if strings.HasSuffix(k, "/") {
			sub, err := listSecretsRecursive(addr, token, mount, full)
			if err == nil {
				all = append(all, sub...)
			}
		} else {
			all = append(all, full)
		}
	}
	return all, nil
}

func readSecretData(addr, token, mount, path string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", addr, mount, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("read secret: status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Data.Data, nil
}
