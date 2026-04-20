package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockEnginesServer(t *testing.T, status int, payload interface{}) *httptest.Server {
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

func TestListSecretEngines_Success(t *testing.T) {
	payload := map[string]interface{}{
		"secret/": map[string]interface{}{
			"type":        "kv",
			"description": "key/value store",
			"options":     map[string]string{"version": "2"},
		},
		"pki/": map[string]interface{}{
			"type":        "pki",
			"description": "PKI engine",
			"options":     map[string]string{},
		},
	}
	srv := mockEnginesServer(t, http.StatusOK, payload)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	mounts, err := client.ListSecretEngines()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) != 2 {
		t.Errorf("expected 2 mounts, got %d", len(mounts))
	}
}

func TestListSecretEngines_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	_, err := client.ListSecretEngines()
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestListSecretEngines_ServerError(t *testing.T) {
	srv := mockEnginesServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	_, err := client.ListSecretEngines()
	if err == nil {
		t.Fatal("expected error for server error")
	}
}
