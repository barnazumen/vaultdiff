package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockLeaseServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.URL.Path == "/v1/secret/missing" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"lease_id":       "secret/data/myapp/abc123",
			"renewable":      true,
			"lease_duration": 3600,
		})
	}))
}

func TestReadLeaseInfo_Success(t *testing.T) {
	srv := mockLeaseServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	info, err := client.ReadLeaseInfo("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.LeaseID != "secret/data/myapp/abc123" {
		t.Errorf("expected lease ID, got %q", info.LeaseID)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
	if info.LeaseDuration.Seconds() != 3600 {
		t.Errorf("expected 3600s duration, got %v", info.LeaseDuration)
	}
}

func TestReadLeaseInfo_NotFound(t *testing.T) {
	srv := mockLeaseServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	_, err := client.ReadLeaseInfo("secret/missing")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestReadLeaseInfo_InvalidToken(t *testing.T) {
	srv := mockLeaseServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "bad-token")
	_, err := client.ReadLeaseInfo("secret/data/myapp")
	if err == nil {
		t.Fatal("expected permission denied error")
	}
}
