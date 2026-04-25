package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockTouchServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/v1/secret/data/missing" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "value"},
				},
			})
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"version": 3},
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestTouchSecret_Success(t *testing.T) {
	srv := mockTouchServer(t)
	defer srv.Close()

	result, err := TouchSecret(srv.URL, "test-token", "secret", "myapp/config")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Path != "myapp/config" {
		t.Errorf("expected path myapp/config, got %s", result.Path)
	}
	if result.Version != 3 {
		t.Errorf("expected version 3, got %d", result.Version)
	}
	if result.Mount != "secret" {
		t.Errorf("expected mount secret, got %s", result.Mount)
	}
}

func TestTouchSecret_NotFound(t *testing.T) {
	srv := mockTouchServer(t)
	defer srv.Close()

	_, err := TouchSecret(srv.URL, "test-token", "secret", "missing")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}

func TestTouchSecret_InvalidToken(t *testing.T) {
	srv := mockTouchServer(t)
	defer srv.Close()

	_, err := TouchSecret(srv.URL, "bad-token", "secret", "myapp/config")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
