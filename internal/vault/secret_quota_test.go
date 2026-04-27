package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockQuotaServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.URL.Path == "/v1/sys/quotas/rate-limit/my-quota" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"path":             "secret/",
					"type":             "rate-limit",
					"max_leases":       100,
					"current_leases":   42,
					"rate":             50.0,
					"burst":            75,
					"interval_seconds": 60,
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestReadSecretQuota_Success(t *testing.T) {
	srv := mockQuotaServer(t)
	defer srv.Close()

	info, err := ReadSecretQuota(srv.URL, "test-token", "my-quota")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "rate-limit" {
		t.Errorf("expected type 'rate-limit', got %q", info.Type)
	}
	if info.CurrentLeases != 42 {
		t.Errorf("expected current_leases 42, got %d", info.CurrentLeases)
	}
	if info.Rate != 50.0 {
		t.Errorf("expected rate 50.0, got %f", info.Rate)
	}
}

func TestReadSecretQuota_NotFound(t *testing.T) {
	srv := mockQuotaServer(t)
	defer srv.Close()

	_, err := ReadSecretQuota(srv.URL, "test-token", "missing-quota")
	if err == nil {
		t.Fatal("expected error for missing quota, got nil")
	}
}

func TestReadSecretQuota_InvalidToken(t *testing.T) {
	srv := mockQuotaServer(t)
	defer srv.Close()

	_, err := ReadSecretQuota(srv.URL, "bad-token", "my-quota")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
