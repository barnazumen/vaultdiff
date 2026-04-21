package vault

import (
	"context"
	"fmt"
	"net/http"
)

// MoveResult holds the outcome of a secret move operation.
type MoveResult struct {
	Source      string
	Destination string
	Versions    int
}

// MoveSecret copies a secret from src to dst and then deletes the source.
// It uses the provided Vault token for authentication.
func MoveSecret(ctx context.Context, addr, token, mount, src, dst string) (*MoveResult, error) {
	client, err := NewClient(addr, token)
	if err != nil {
		return nil, fmt.Errorf("move secret: create client: %w", err)
	}

	// Copy source to destination first
	result, err := CopySecret(ctx, client, mount, src, dst)
	if err != nil {
		return nil, fmt.Errorf("move secret: copy phase: %w", err)
	}

	// Delete the source secret after successful copy
	delURL := fmt.Sprintf("%s/v1/%s/metadata/%s", addr, mount, src)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, delURL, nil)
	if err != nil {
		return nil, fmt.Errorf("move secret: build delete request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("move secret: delete source: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("move secret: delete source: permission denied (status %d)", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("move secret: delete source: unexpected status %d", resp.StatusCode)
	}

	return &MoveResult{
		Source:      src,
		Destination: dst,
		Versions:    result.Versions,
	}, nil
}
