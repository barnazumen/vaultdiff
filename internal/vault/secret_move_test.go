package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockMoveServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		path := r.URL.Path

		// Handle metadata GET for listing versions (copy phase)
		if r.Method == http.MethodGet && strings.Contains(path, "/metadata/") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"versions": map[string]interface{}{
						"1": map[string]interface{}{"destroyed": false, "deletion_time": ""},
					},
				},
			})
			return
		}

		// Handle data GET for reading secret data (copy phase)
		if r.Method == http.MethodGet && strings.Contains(path, "/data/") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "value"},
				},
			})
			return
		}

		// Handle data POST for writing secret (copy phase)
		if r.Method == http.MethodPost && strings.Contains(path, "/data/") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 1}})
			return
		}

		// Handle metadata DELETE for removing source (move phase)
		if r.Method == http.MethodDelete && strings.Contains(path, "/metadata/") {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestMoveSecret_Success(t *testing.T) {
	srv := mockMoveServer(t)
	defer srv.Close()

	result, err := MoveSecret(context.Background(), srv.URL, "test-token", "secret", "src/key", "dst/key")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Source != "src/key" {
		t.Errorf("expected source src/key, got %s", result.Source)
	}
	if result.Destination != "dst/key" {
		t.Errorf("expected destination dst/key, got %s", result.Destination)
	}
}

func TestMoveSecret_InvalidToken(t *testing.T) {
	srv := mockMoveServer(t)
	defer srv.Close()

	_, err := MoveSecret(context.Background(), srv.URL, "bad-token", "secret", "src/key", "dst/key")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
