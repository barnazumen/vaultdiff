package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockNamespaceServer(t *testing.T, token string, statusCode int, payload interface{}) *httptest.Server {
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

func TestListNamespaces_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"team-a/": map[string]interface{}{
				"id":              "abc123",
				"custom_metadata": map[string]string{"env": "prod"},
			},
			"team-b/": map[string]interface{}{
				"id":              "def456",
				"custom_metadata": map[string]string{},
			},
		},
	}

	srv := mockNamespaceServer(t, "test-token", http.StatusOK, payload)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	ns, err := client.ListNamespaces("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ns) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(ns))
	}
}

func TestListNamespaces_NotFound(t *testing.T) {
	srv := mockNamespaceServer(t, "test-token", http.StatusNotFound, nil)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	_, err := client.ListNamespaces("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
}

func TestListNamespaces_InvalidToken(t *testing.T) {
	srv := mockNamespaceServer(t, "valid-token", http.StatusOK, nil)
	defer srv.Close()

	client := NewClient(srv.URL, "wrong-token")
	_, err := client.ListNamespaces("")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
