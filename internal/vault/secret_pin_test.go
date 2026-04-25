package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockPinServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if strings.Contains(r.URL.Path, "missing") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusNoContent)
	}))
}

func TestPinSecret_Success(t *testing.T) {
	srv := mockPinServer(t)
	defer srv.Close()

	result, err := PinSecret(srv.URL, "test-token", "secret", "myapp/config", 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result.Pinned {
		t.Error("expected Pinned to be true")
	}
	if result.Version != 3 {
		t.Errorf("expected version 3, got %d", result.Version)
	}
	if result.Path != "myapp/config" {
		t.Errorf("unexpected path: %s", result.Path)
	}
}

func TestPinSecret_NotFound(t *testing.T) {
	srv := mockPinServer(t)
	defer srv.Close()

	_, err := PinSecret(srv.URL, "test-token", "secret", "missing/secret", 1)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPinSecret_InvalidToken(t *testing.T) {
	srv := mockPinServer(t)
	defer srv.Close()

	_, err := PinSecret(srv.URL, "bad-token", "secret", "myapp/config", 2)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUnpinSecret_Success(t *testing.T) {
	srv := mockPinServer(t)
	defer srv.Close()

	result, err := UnpinSecret(srv.URL, "test-token", "secret", "myapp/config")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Pinned {
		t.Error("expected Pinned to be false after unpin")
	}
	if result.Version != 0 {
		t.Errorf("expected version 0 after unpin, got %d", result.Version)
	}
}

func TestUnpinSecret_InvalidToken(t *testing.T) {
	srv := mockPinServer(t)
	defer srv.Close()

	_, err := UnpinSecret(srv.URL, "bad-token", "secret", "myapp/config")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Errorf("unexpected error: %v", err)
	}
}
