package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockTokenRenewServer(t *testing.T, statusCode int, payload interface{}) *httptest.Server {
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

func TestRenewToken_Success(t *testing.T) {
	payload := map[string]interface{}{
		"auth": map[string]interface{}{
			"client_token":   "new-token-abc",
			"lease_duration": 3600,
			"renewable":      true,
		},
	}
	server := mockTokenRenewServer(t, http.StatusOK, payload)
	defer server.Close()

	client := NewClient(server.URL, "valid-token")
	result, err := client.RenewToken("valid-token")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.ClientToken != "new-token-abc" {
		t.Errorf("expected client_token 'new-token-abc', got '%s'", result.ClientToken)
	}
	if result.LeaseDuration != 3600 {
		t.Errorf("expected lease_duration 3600, got %d", result.LeaseDuration)
	}
	if !result.Renewable {
		t.Error("expected renewable to be true")
	}
}

func TestRenewToken_InvalidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token")
	_, err := client.RenewToken("bad-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestRenewToken_ServerError(t *testing.T) {
	server := mockTokenRenewServer(t, http.StatusInternalServerError, nil)
	defer server.Close()

	client := NewClient(server.URL, "valid-token")
	_, err := client.RenewToken("valid-token")
	if err == nil {
		t.Fatal("expected error on server error, got nil")
	}
}
