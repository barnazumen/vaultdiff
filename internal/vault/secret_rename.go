package vault

import (
	"context"
	"fmt"
)

// RenameSecret copies a secret from srcPath to dstPath and deletes the source.
// Both paths are under the same mount. The latest version data is used.
func (c *Client) RenameSecret(ctx context.Context, mount, srcPath, dstPath string) error {
	data, err := c.ReadSecretVersion(ctx, mount, srcPath, 0)
	if err != nil {
		return fmt.Errorf("rename: read source %q: %w", srcPath, err)
	}

	writePath := fmt.Sprintf("%s/data/%s", mount, dstPath)
	body := map[string]any{
		"data": data,
	}

	resp, err := c.write(ctx, writePath, body)
	if err != nil {
		return fmt.Errorf("rename: write destination %q: %w", dstPath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("rename: unexpected status %d writing %q", resp.StatusCode, dstPath)
	}

	if err := c.DeleteSecret(ctx, mount, srcPath); err != nil {
		return fmt.Errorf("rename: delete source %q after copy: %w", srcPath, err)
	}

	return nil
}

// DeleteSecret permanently deletes all versions and metadata for a KV v2 secret.
func (c *Client) DeleteSecret(ctx context.Context, mount, path string) error {
	metaPath := fmt.Sprintf("%s/metadata/%s", mount, path)
	resp, err := c.delete(ctx, metaPath)
	if err != nil {
		return fmt.Errorf("delete secret %q: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("delete secret %q: unexpected status %d", path, resp.StatusCode)
	}
	return nil
}
