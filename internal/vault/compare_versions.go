package vault

import "fmt"

// VersionPair holds two version numbers to compare.
type VersionPair struct {
	From int
	To   int
}

// ReadVersionPair fetches two versions of a secret and returns their data maps.
func (c *Client) ReadVersionPair(mountPath, secretPath string, pair VersionPair) (map[string]interface{}, map[string]interface{}, error) {
	from, err := c.ReadSecretVersion(mountPath, secretPath, pair.From)
	if err != nil {
		return nil, nil, fmt.Errorf("reading version %d: %w", pair.From, err)
	}

	to, err := c.ReadSecretVersion(mountPath, secretPath, pair.To)
	if err != nil {
		return nil, nil, fmt.Errorf("reading version %d: %w", pair.To, err)
	}

	return from, to, nil
}

// ResolveVersionPair resolves "latest" semantics: if To is 0, use the highest
// available version; if From is 0, use To-1.
func ResolveVersionPair(versions []int, pair VersionPair) (VersionPair, error) {
	if len(versions) == 0 {
		return pair, fmt.Errorf("no versions available")
	}

	max := versions[len(versions)-1]

	if pair.To == 0 {
		pair.To = max
	}
	if pair.From == 0 {
		pair.From = pair.To - 1
	}

	if pair.From < 1 {
		return pair, fmt.Errorf("resolved 'from' version %d is invalid", pair.From)
	}
	if pair.From >= pair.To {
		return pair, fmt.Errorf("'from' version %d must be less than 'to' version %d", pair.From, pair.To)
	}

	return pair, nil
}
