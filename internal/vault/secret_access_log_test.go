package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockAccessLogServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(status)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestReadAccessLog_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"versions": map[string]interface{}{
				"1": map[string]interface{}{
					"created_time":  "2024-01-01T10:00:00Z",
					"deletion_time": "",
					"destroyed":     false,
				},
				"2": map[string]interface{}{
					"created_time":  "2024-01-02T10:00:00Z",
					"deletion_time": "",
					"destroyed":     false,
				},
			},
		},
	}
	ts := mockAccessLogServer(t, http.StatusOK, payload)
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	result, err := c.ReadAccessLog("secret", "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Path != "myapp/config" {
		t.Errorf("expected path 'myapp/config', got %q", result.Path)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
}

func TestReadAccessLog_NotFound(t *testing.T) {
	ts := mockAccessLogServer(t, http.StatusNotFound, nil)
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	_, err := c.ReadAccessLog("secret", "missing/path")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestReadAccessLog_InvalidToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "bad-token")
	_, err := c.ReadAccessLog("secret", "myapp/config")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
