package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockAuditTrailServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.URL.Path == "/v1/secret/metadata/missing" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"created_time": "2024-01-01T00:00:00Z",
				"versions": map[string]interface{}{
					"1": map[string]interface{}{
						"created_time":  "2024-01-01T00:00:00Z",
						"deletion_time": "",
						"destroyed":     false,
					},
					"2": map[string]interface{}{
						"created_time":  "2024-02-01T00:00:00Z",
						"deletion_time": "2024-03-01T00:00:00Z",
						"destroyed":     false,
					},
				},
			},
		})
	}))
}

func TestReadAuditTrail_Success(t *testing.T) {
	srv := mockAuditTrailServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	trail, err := client.ReadAuditTrail("secret", "myapp/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trail.Path != "myapp/config" {
		t.Errorf("expected path myapp/config, got %s", trail.Path)
	}
	if len(trail.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(trail.Entries))
	}
}

func TestReadAuditTrail_NotFound(t *testing.T) {
	srv := mockAuditTrailServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	_, err := client.ReadAuditTrail("secret", "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadAuditTrail_InvalidToken(t *testing.T) {
	srv := mockAuditTrailServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "bad-token")
	_, err := client.ReadAuditTrail("secret", "myapp/config")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
