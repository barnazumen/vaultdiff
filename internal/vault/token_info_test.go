package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockTokenInfoServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestGetTokenInfo_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":     "abc123",
			"display_name": "token-test",
			"policies":     []string{"default", "admin"},
			"ttl":          3600,
			"renewable":    true,
			"entity_id":    "ent-xyz",
		},
	}
	srv := mockTokenInfoServer(t, http.StatusOK, payload)
	defer srv.Close()

	client := NewClient(srv.URL, "valid-token")
	info, err := client.GetTokenInfo()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if info.DisplayName != "token-test" {
		t.Errorf("expected display_name token-test, got %s", info.DisplayName)
	}
	if len(info.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(info.Policies))
	}
	if info.TTL != 3600 {
		t.Errorf("expected TTL 3600, got %d", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
}

func TestGetTokenInfo_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "bad-token")
	_, err := client.GetTokenInfo()
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestGetTokenInfo_ServerError(t *testing.T) {
	srv := mockTokenInfoServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	client := NewClient(srv.URL, "valid-token")
	_, err := client.GetTokenInfo()
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}
