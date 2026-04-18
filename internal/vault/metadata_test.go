package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReadSecretMetadata_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"current_version": 3,
				"oldest_version":  1,
				"created_time":    time.Now().Format(time.RFC3339),
				"versions": map[string]interface{}{
					"1": map[string]interface{}{"version": 1, "destroyed": false},
					"2": map[string]interface{}{"version": 2, "destroyed": false},
					"3": map[string]interface{}{"version": 3, "destroyed": false},
				},
			},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")
	meta, err := client.ReadSecretMetadata("secret", "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.CurrentVersion != 3 {
		t.Errorf("expected current_version 3, got %d", meta.CurrentVersion)
	}
	if len(meta.Versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(meta.Versions))
	}
}

func TestReadSecretMetadata_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")
	_, err := client.ReadSecretMetadata("secret", "missing/path")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestReadSecretMetadata_InvalidToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "bad-token")
	_, err := client.ReadSecretMetadata("secret", "myapp/config")
	if err == nil {
		t.Fatal("expected error for forbidden")
	}
}
