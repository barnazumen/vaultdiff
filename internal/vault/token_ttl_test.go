package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func mockTokenTTLServer(t *testing.T, token string, statusCode int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != token {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(statusCode)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestGetTokenTTL_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"ttl":              3600,
			"creation_ttl":     86400,
			"expire_time":      "2099-01-01T00:00:00Z",
			"explicit_max_ttl": 0,
			"period":           0,
		},
	}
	server := mockTokenTTLServer(t, "test-token", http.StatusOK, payload)
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	info, err := client.GetTokenTTL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TTL != 3600 {
		t.Errorf("expected TTL 3600, got %d", info.TTL)
	}
	if info.CreationTTL != 86400 {
		t.Errorf("expected CreationTTL 86400, got %d", info.CreationTTL)
	}
	expected, _ := time.Parse(time.RFC3339, "2099-01-01T00:00:00Z")
	if !info.ExpireTime.Equal(expected) {
		t.Errorf("expected ExpireTime %v, got %v", expected, info.ExpireTime)
	}
}

func TestGetTokenTTL_InvalidToken(t *testing.T) {
	server := mockTokenTTLServer(t, "valid-token", http.StatusOK, nil)
	defer server.Close()

	client := NewClient(server.URL, "wrong-token")
	_, err := client.GetTokenTTL()
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestGetTokenTTL_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "any-token")
	_, err := client.GetTokenTTL()
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}
