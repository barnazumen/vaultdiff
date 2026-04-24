package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockPromoteServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/data/") {
			if strings.Contains(r.URL.Path, "missing") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"api_key": "secret123", "env": "staging"},
					"metadata": map[string]interface{}{"version": 3},
				},
			})
			return
		}
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestPromoteSecret_Success(t *testing.T) {
	srv := mockPromoteServer(t)
	defer srv.Close()

	result, err := PromoteSecret(srv.URL, "test-token", "secret", "prod-secret", "myapp/config", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SourceMount != "secret" {
		t.Errorf("expected source mount 'secret', got %q", result.SourceMount)
	}
	if result.DestMount != "prod-secret" {
		t.Errorf("expected dest mount 'prod-secret', got %q", result.DestMount)
	}
	if result.Version != 3 {
		t.Errorf("expected version 3, got %d", result.Version)
	}
	if len(result.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result.Keys))
	}
}

func TestPromoteSecret_SourceNotFound(t *testing.T) {
	srv := mockPromoteServer(t)
	defer srv.Close()

	_, err := PromoteSecret(srv.URL, "test-token", "secret", "prod-secret", "missing/path", "")
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestPromoteSecret_InvalidToken(t *testing.T) {
	srv := mockPromoteServer(t)
	defer srv.Close()

	_, err := PromoteSecret(srv.URL, "bad-token", "secret", "prod-secret", "myapp/config", "")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestPromoteSecret_CustomDstPath(t *testing.T) {
	srv := mockPromoteServer(t)
	defer srv.Close()

	result, err := PromoteSecret(srv.URL, "test-token", "secret", "prod-secret", "myapp/config", "myapp/prod-config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Path != "myapp/config" {
		t.Errorf("expected path 'myapp/config', got %q", result.Path)
	}
}
