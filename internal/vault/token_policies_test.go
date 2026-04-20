package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockTokenPoliciesServer(t *testing.T, statusCode int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(statusCode)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestGetTokenPolicies_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"policies":          []string{"default", "admin"},
			"token_policies":    []string{"admin"},
			"identity_policies": []string{"base"},
		},
	}
	server := mockTokenPoliciesServer(t, http.StatusOK, payload)
	defer server.Close()

	client := NewClient(server.URL, "valid-token")
	result, err := client.GetTokenPolicies("some-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(result.Policies))
	}
	if result.Policies[0] != "default" {
		t.Errorf("expected first policy to be 'default', got %q", result.Policies[0])
	}
	if len(result.TokenPolicies) != 1 || result.TokenPolicies[0] != "admin" {
		t.Errorf("expected token_policies [admin], got %v", result.TokenPolicies)
	}
	if len(result.IdentityPolicies) != 1 || result.IdentityPolicies[0] != "base" {
		t.Errorf("expected identity_policies [base], got %v", result.IdentityPolicies)
	}
}

func TestGetTokenPolicies_InvalidToken(t *testing.T) {
	server := mockTokenPoliciesServer(t, http.StatusForbidden, nil)
	defer server.Close()

	client := NewClient(server.URL, "")
	_, err := client.GetTokenPolicies("bad-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestGetTokenPolicies_NotFound(t *testing.T) {
	server := mockTokenPoliciesServer(t, http.StatusNotFound, nil)
	defer server.Close()

	client := NewClient(server.URL, "valid-token")
	_, err := client.GetTokenPolicies("missing-token")
	if err == nil {
		t.Fatal("expected error for not found token, got nil")
	}
}

func TestGetTokenPolicies_ServerError(t *testing.T) {
	server := mockTokenPoliciesServer(t, http.StatusInternalServerError, nil)
	defer server.Close()

	client := NewClient(server.URL, "valid-token")
	_, err := client.GetTokenPolicies("some-token")
	if err == nil {
		t.Fatal("expected error on server error, got nil")
	}
}
