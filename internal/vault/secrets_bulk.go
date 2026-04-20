package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// BulkSecretResult holds the result for a single secret path in a bulk read.
type BulkSecretResult struct {
	Path   string
	Data   map[string]interface{}
	Error  error
}

// ReadSecretsbulk reads multiple secret paths concurrently and returns
// a slice of BulkSecretResult, one per path.
func (c *Client) ReadSecretsBulk(mount string, paths []string, version int) []BulkSecretResult {
	type work struct {
		index int
		path  string
	}

	results := make([]BulkSecretResult, len(paths))
	ch := make(chan work, len(paths))

	for i, p := range paths {
		ch <- work{i, p}
	}
	close(ch)

	const workers = 5
	done := make(chan struct{}, workers)

	for w := 0; w < workers; w++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for job := range ch {
				data, err := c.readSingleSecret(mount, job.path, version)
				results[job.index] = BulkSecretResult{
					Path:  job.path,
					Data:  data,
					Error: err,
				}
			}
		}()
	}

	for w := 0; w < workers; w++ {
		<-done
	}

	return results
}

func (c *Client) readSingleSecret(mount, path string, version int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", c.Address, mount, path)
	if version > 0 {
		url = fmt.Sprintf("%s?version=%d", url, version)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for path %s", resp.StatusCode, path)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return result.Data.Data, nil
}
