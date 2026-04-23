package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockSearchServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch r.URL.Path {
		case "/v1/secret/metadata/":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"db/", "api-key"},
				},
			})
		case "/v1/secret/metadata/db/":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"postgres"},
				},
			})
		case "/v1/secret/data/db/postgres":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"password": "supersecret", "user": "admin"},
				},
			})
		case "/v1/secret/data/api-key":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "abc123"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestSearchSecrets_MatchByKey(t *testing.T) {
	srv := mockSearchServer(t)
	defer srv.Close()

	results, err := SearchSecrets(srv.URL, "test-token", "secret", "", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "db/postgres" {
		t.Errorf("expected path db/postgres, got %s", results[0].Path)
	}
}

func TestSearchSecrets_MatchByValue(t *testing.T) {
	srv := mockSearchServer(t)
	defer srv.Close()

	results, err := SearchSecrets(srv.URL, "test-token", "secret", "", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "api-key" {
		t.Errorf("expected path api-key, got %s", results[0].Path)
	}
}

func TestSearchSecrets_NoMatch(t *testing.T) {
	srv := mockSearchServer(t)
	defer srv.Close()

	results, err := SearchSecrets(srv.URL, "test-token", "secret", "", "notexist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchSecrets_InvalidToken(t *testing.T) {
	srv := mockSearchServer(t)
	defer srv.Close()

	_, err := SearchSecrets(srv.URL, "bad-token", "secret", "", "password")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
