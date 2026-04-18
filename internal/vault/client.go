package vault

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods.
type Client struct {
	vc *vaultapi.Client
}

// NewClient creates a new Vault client using the provided address and token.
// If addr or token are empty, they fall back to VAULT_ADDR / VAULT_TOKEN env vars.
func NewClient(addr, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	if addr != "" {
		cfg.Address = addr
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	if token != "" {
		vc.SetToken(token)
	}

	return &Client{vc: vc}, nil
}

// ReadSecretVersion reads a specific version of a KV v2 secret.
// mount is the KV mount path (e.g. "secret"), secretPath is the key path.
// Use version 0 to read the latest version.
func (c *Client) ReadSecretVersion(mount, secretPath string, version int) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/data/%s", mount, secretPath)

	params := map[string][]string{}
	if version > 0 {
		params["version"] = []string{fmt.Sprintf("%d", version)}
	}

	secret, err := c.vc.Logical().ReadWithData(path, params)
	if err != nil {
		return nil, fmt.Errorf("reading secret %q version %d: %w", secretPath, version, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret %q version %d not found", secretPath, version)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format for secret %q", secretPath)
	}

	return data, nil
}

// ListSecrets returns the list of keys under the given path in a KV v2 mount.
func (c *Client) ListSecrets(mount, secretPath string) ([]string, error) {
	path := fmt.Sprintf("%s/metadata/%s", mount, secretPath)

	secret, err := c.vc.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("listing secrets at %q: %w", secretPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secrets found at %q", secretPath)
	}

	raw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys format at %q", secretPath)
	}

	keys := make([]string, len(raw))
	for i, k := range raw {
		keys[i] = fmt.Sprintf("%v", k)
	}
	return keys, nil
}
