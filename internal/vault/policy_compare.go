package vault

import "fmt"

// PolicyCompareResult holds the result of comparing two Vault policies.
type PolicyCompareResult struct {
	Path   string
	FromHCL string
	ToHCL   string
	Diff   []PolicyDiffLine
}

// ComparePolicies fetches two versions of a policy by name and diffs them.
func ComparePolicies(client *Client, policyName, fromToken, toToken string) (*PolicyCompareResult, error) {
	fromHCL, err := readPolicyWithToken(client, policyName, fromToken)
	if err != nil {
		return nil, fmt.Errorf("reading 'from' policy: %w", err)
	}

	toHCL, err := readPolicyWithToken(client, policyName, toToken)
	if err != nil {
		return nil, fmt.Errorf("reading 'to' policy: %w", err)
	}

	diff := DiffPolicies(fromHCL, toHCL)

	return &PolicyCompareResult{
		Path:    policyName,
		FromHCL: fromHCL,
		ToHCL:   toHCL,
		Diff:    diff,
	}, nil
}

// readPolicyWithToken temporarily overrides the client token and reads a policy.
func readPolicyWithToken(client *Client, policyName, token string) (string, error) {
	original := client.Token
	if token != "" {
		client.Token = token
	}
	defer func() { client.Token = original }()

	return ReadPolicy(client, policyName)
}
