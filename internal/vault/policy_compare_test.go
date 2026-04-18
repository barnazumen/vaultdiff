package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockPolicyCompareServer(t *testing.T, policyName, hcl string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, policyName) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]string{"rules": hcl},
		})
	}))
}

func TestComparePolicies_NoDiff(t *testing.T) {
	hcl := `path "secret/*" { capabilities = ["read"] }`
	server := mockPolicyCompareServer(t, "my-policy", hcl)
	defer server.Close()

	client := &Client{Address: server.URL, Token: "test-token"}
	result, err := ComparePolicies(client, "my-policy", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Path != "my-policy" {
		t.Errorf("expected path 'my-policy', got %q", result.Path)
	}
	for _, line := range result.Diff {
		if line.Type == "added" || line.Type == "removed" {
			t.Errorf("expected no changes, got: %+v", line)
		}
	}
}

func TestComparePolicies_WithDiff(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hcl := `path "secret/a" { capabilities = ["read"] }`
		if callCount > 0 {
			hcl = `path "secret/b" { capabilities = ["read"] }`
		}
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]string{"rules": hcl},
		})
	}))
	defer server.Close()

	client := &Client{Address: server.URL, Token: "test-token"}
	result, err := ComparePolicies(client, "my-policy", "", "other-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Diff) == 0 {
		t.Error("expected non-empty diff")
	}
}
