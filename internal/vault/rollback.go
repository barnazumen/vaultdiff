package vault

import (
	"fmt"
	"net/http"
)

// RollbackResult holds the result of a rollback operation.
type RollbackResult struct {
	Path       string
	FromVersion int
	ToVersion   int
	Success    bool
}

// RollbackToVersion rolls back a KV v2 secret to a specific version by
// re-writing that version's data as the new current version.
func (c *Client) RollbackToVersion(mountPath, secretPath string, targetVersion int) (*RollbackResult, error) {
	secret, err := c.ReadSecretVersion(mountPath, secretPath, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("rollback: read version %d: %w", targetVersion, err)
	}

	writePath := fmt.Sprintf("%s/data/%s", mountPath, secretPath)
	payload := map[string]interface{}{
		"data": secret,
	}

	resp, err := c.rawPost(writeePath, payload)
	if err != nil {
		return nil, fmt.Errorf("rollback: write: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rollback: unexpected status %d", resp.StatusCode)
	}

	versions, err := c.ListVersions(mountPath, secretPath)
	if err != nil {
		return nil, fmt.Errorf("rollback: list versions after write: %w", err)
	}

	newVersion := len(versions)
	return &RollbackResult{
		Path:        secretPath,
		FromVersion: newVersion - 1,
		ToVersion:   newVersion,
		Success:     true,
	}, nil
}
