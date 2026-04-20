package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockCapabilitiesServer(t *testing.T, token string, caps []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != token {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"capabilities": caps,
		})
	}))
}

func TestGetTokenCapabilities_Success(t *testing.T) {
	expected := []string{"read", "list"}
	server := mockCapabilitiesServer(t, "test-token", expected)
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	caps, err := client.GetTokenCapabilities("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(caps.Capabilities) != len(expected) {
		t.Fatalf("expected %d capabilities, got %d", len(expected), len(caps.Capabilities))
	}
	for i, c := range caps.Capabilities {
		if c != expected[i] {
			t.Errorf("capability[%d]: expected %q, got %q", i, expected[i], c)
		}
	}
	if caps.Path != "secret/data/myapp" {
		t.Errorf("expected path %q, got %q", "secret/data/myapp", caps.Path)
	}
}

func TestGetTokenCapabilities_InvalidToken(t *testing.T) {
	server := mockCapabilitiesServer(t, "valid-token", []string{"read"})
	defer server.Close()

	client := NewClient(server.URL, "wrong-token")
	_, err := client.GetTokenCapabilities("secret/data/myapp")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestGetTokenCapabilities_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "any-token")
	_, err := client.GetTokenCapabilities("secret/data/myapp")
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}
