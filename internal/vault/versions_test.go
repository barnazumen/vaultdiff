package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListVersions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"versions": map[string]interface{}{
					"1": map[string]interface{}{
						"created_time": "2024-01-01T00:00:00Z",
						"deletion_time": "",
						"destroyed":    false,
					},
					"2": map[string]interface{}{
						"created_time": "2024-01-02T00:00:00Z",
						"deletion_time": "",
						"destroyed":    false,
					},
				},
			},
		})
	}))
	defer server.Close()

	c := &Client{Address: server.URL, Token: "test-token", HTTP: server.Client()}
	versions, err := c.ListVersions("secret", "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}

func TestListVersions_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := &Client{Address: server.URL, Token: "test-token", HTTP: server.Client()}
	_, err := c.ListVersions("secret", "missing/path")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestListVersions_InvalidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c := &Client{Address: server.URL, Token: "bad-token", HTTP: server.Client()}
	_, err := c.ListVersions("secret", "myapp/config")
	if err == nil {
		t.Fatal("expected error for forbidden")
	}
}
