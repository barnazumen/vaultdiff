package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockRenameServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/old-key":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"data": map[string]any{"username": "alice"},
					"metadata": map[string]any{"version": 1},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/secret/data/new-key":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/secret/metadata/old-key":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestRenameSecret_Success(t *testing.T) {
	srv := mockRenameServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	err := c.RenameSecret(context.Background(), "secret", "old-key", "new-key")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRenameSecret_SourceNotFound(t *testing.T) {
	srv := mockRenameServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	err := c.RenameSecret(context.Background(), "secret", "missing-key", "new-key")
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestRenameSecret_InvalidToken(t *testing.T) {
	srv := mockRenameServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "bad-token")
	err := c.RenameSecret(context.Background(), "secret", "old-key", "new-key")
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}
}
