package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockPolicyServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.URL.Path == "/v1/sys/policies/acl/my-policy" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]string{
					"name":  "my-policy",
					"rules": `path "secret/*" { capabilities = ["read"] }`,
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestReadPolicy_Success(t *testing.T) {
	server := mockPolicyServer(t)
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	policy, err := client.ReadPolicy("my-policy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Name != "my-policy" {
		t.Errorf("expected name 'my-policy', got %q", policy.Name)
	}
	if policy.Rules == "" {
		t.Error("expected non-empty rules")
	}
}

func TestReadPolicy_NotFound(t *testing.T) {
	server := mockPolicyServer(t)
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	_, err := client.ReadPolicy("missing-policy")
	if err == nil {
		t.Fatal("expected error for missing policy")
	}
}

func TestReadPolicy_InvalidToken(t *testing.T) {
	server := mockPolicyServer(t)
	defer server.Close()

	client := NewClient(server.URL, "bad-token")
	_, err := client.ReadPolicy("my-policy")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
