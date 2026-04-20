package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockAuthServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(status)
		if payload != nil {
			json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestLookupToken_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":    "abc123",
			"policies":    []string{"default", "admin"},
			"ttl":         3600,
			"expire_time": "2099-01-01T00:00:00Z",
			"display_name": "token-test",
		},
	}
	ts := mockAuthServer(t, http.StatusOK, payload)
	defer ts.Close()

	c := NewClient(ts.URL, "valid-token")
	info, err := c.LookupToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if len(info.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(info.Policies))
	}
}

func TestLookupToken_InvalidToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "bad-token")
	_, err := c.LookupToken()
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestLookupToken_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "some-token")
	_, err := c.LookupToken()
	if err == nil {
		t.Fatal("expected error for server error")
	}
}
