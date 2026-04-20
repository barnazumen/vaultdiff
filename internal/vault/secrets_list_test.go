package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockSecretsListServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.Method != "LIST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/v1/secret/metadata/myapp":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"db-password", "api-key", "tls-cert"},
				},
			})
		case "/v1/secret/metadata/empty":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestListSecrets_Success(t *testing.T) {
	srv := mockSecretsListServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	result, err := client.ListSecrets("secret", "myapp")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(result.Keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(result.Keys))
	}
	if result.Keys[0] != "db-password" {
		t.Errorf("expected first key to be 'db-password', got %q", result.Keys[0])
	}
}

func TestListSecrets_NotFound(t *testing.T) {
	srv := mockSecretsListServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	_, err := client.ListSecrets("secret", "empty")
	if err == nil {
		t.Fatal("expected error for not found path, got nil")
	}
}

func TestListSecrets_InvalidToken(t *testing.T) {
	srv := mockSecretsListServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "bad-token")
	_, err := client.ListSecrets("secret", "myapp")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
