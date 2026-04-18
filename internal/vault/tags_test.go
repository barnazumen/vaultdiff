package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadSecretTags_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"custom_metadata": map[string]string{
					"owner": "team-a",
					"env":   "production",
				},
			},
		})
	}))
	defer server.Close()

	client := &Client{Address: server.URL, Token: "test-token", HTTP: server.Client()}
	tags, err := client.ReadSecretTags("secret", "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tags["owner"] != "team-a" {
		t.Errorf("expected owner=team-a, got %s", tags["owner"])
	}
	if tags["env"] != "production" {
		t.Errorf("expected env=production, got %s", tags["env"])
	}
}

func TestReadSecretTags_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{Address: server.URL, Token: "test-token", HTTP: server.Client()}
	_, err := client.ReadSecretTags("secret", "missing/path")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadSecretTags_EmptyMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"custom_metadata": nil,
			},
		})
	}))
	defer server.Close()

	client := &Client{Address: server.URL, Token: "test-token", HTTP: server.Client()}
	tags, err := client.ReadSecretTags("secret", "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected empty tags, got %v", tags)
	}
}
