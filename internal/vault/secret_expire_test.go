package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func mockExpireServer(t *testing.T, token string, createdTime string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != token {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.URL.Path == "/v1/secret/metadata/myapp/db" {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"data": map[string]interface{}{
					"current_version": 2,
					"versions": map[string]interface{}{
						"2": map[string]string{"created_time": createdTime},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestCheckSecretExpiry_NotExpired(t *testing.T) {
	created := time.Now().Add(-5 * 24 * time.Hour).UTC().Format(time.RFC3339Nano)
	ts := mockExpireServer(t, "test-token", created)
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	result, err := c.CheckSecretExpiry("secret", "myapp/db", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Expired {
		t.Errorf("expected not expired, got expired")
	}
	if result.Version != 2 {
		t.Errorf("expected version 2, got %d", result.Version)
	}
}

func TestCheckSecretExpiry_Expired(t *testing.T) {
	created := time.Now().Add(-60 * 24 * time.Hour).UTC().Format(time.RFC3339Nano)
	ts := mockExpireServer(t, "test-token", created)
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	result, err := c.CheckSecretExpiry("secret", "myapp/db", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Expired {
		t.Errorf("expected expired, got not expired")
	}
}

func TestCheckSecretExpiry_InvalidToken(t *testing.T) {
	created := time.Now().UTC().Format(time.RFC3339Nano)
	ts := mockExpireServer(t, "valid-token", created)
	defer ts.Close()

	c := NewClient(ts.URL, "bad-token")
	_, err := c.CheckSecretExpiry("secret", "myapp/db", 30)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestCheckSecretExpiry_NotFound(t *testing.T) {
	created := time.Now().UTC().Format(time.RFC3339Nano)
	ts := mockExpireServer(t, "test-token", created)
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")
	_, err := c.CheckSecretExpiry("secret", "nonexistent/path", 30)
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}
