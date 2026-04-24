package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockLockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/metadata/") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"custom_metadata": map[string]string{
						"locked":      "true",
						"locked_by":   "test-token",
						"locked_at":   "2024-01-01T00:00:00Z",
						"lock_reason": "maintenance",
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestLockSecret_Success(t *testing.T) {
	srv := mockLockServer(t)
	defer srv.Close()
	c := NewClient(srv.URL)
	info, err := c.LockSecret("myapp/db", "secret", "test-token", "maintenance", 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected lock info, got nil")
	}
	if info.Reason != "maintenance" {
		t.Errorf("expected reason 'maintenance', got %q", info.Reason)
	}
	if info.ExpiresAt.IsZero() {
		t.Error("expected non-zero ExpiresAt for ttl>0")
	}
}

func TestUnlockSecret_Success(t *testing.T) {
	srv := mockLockServer(t)
	defer srv.Close()
	c := NewClient(srv.URL)
	if err := c.UnlockSecret("myapp/db", "secret", "test-token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsLocked_True(t *testing.T) {
	srv := mockLockServer(t)
	defer srv.Close()
	c := NewClient(srv.URL)
	locked, info, err := c.IsLocked("myapp/db", "secret", "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !locked {
		t.Error("expected secret to be locked")
	}
	if info == nil || info.LockedBy != "test-token" {
		t.Errorf("unexpected lock info: %+v", info)
	}
}

func TestLockSecret_InvalidToken(t *testing.T) {
	srv := mockLockServer(t)
	defer srv.Close()
	c := NewClient(srv.URL)
	_, err := c.LockSecret("myapp/db", "secret", "bad-token", "test", 0)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
