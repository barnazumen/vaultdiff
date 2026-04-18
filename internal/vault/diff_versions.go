package vault

import (
	"fmt"

	"github.com/your-org/vaultdiff/internal/diff"
)

// DiffVersions fetches two versions of a secret and returns a list of Changes.
func DiffVersions(client *Client, mountPath, secretPath string, versionA, versionB int) ([]diff.Change, error) {
	secretA, err := client.ReadSecretVersion(mountPath, secretPath, versionA)
	if err != nil {
		return nil, fmt.Errorf("reading version %d: %w", versionA, err)
	}

	secretB, err := client.ReadSecretVersion(mountPath, secretPath, versionB)
	if err != nil {
		return nil, fmt.Errorf("reading version %d: %w", versionB, err)
	}

	changes := diff.Compare(secretA, secretB)
	return changes, nil
}
