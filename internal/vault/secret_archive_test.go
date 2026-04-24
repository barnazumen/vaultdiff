package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockArchiveServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(statusCode)
	}))
}

func TestArchiveSecretVersion_Success(t *testing.T) {
	ts := mockArchiveServer(t, http.StatusNoContent)
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")
	result, err := client.ArchiveSecretVersion("secret", "myapp/config", 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result.Archived {
		t.Error("expected Archived to be true")
	}
	if result.Version != 3 {
		t.Errorf("expected version 3, got %d", result.Version)
	}
	if result.Path != "myapp/config" {
		t.Errorf("expected path myapp/config, got %s", result.Path)
	}
}

func TestArchiveSecretVersion_NotFound(t *testing.T) {
	ts := mockArchiveServer(t, http.StatusNotFound)
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")
	_, err := client.ArchiveSecretVersion("secret", "missing/path", 1)
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
}

func TestArchiveSecretVersion_InvalidToken(t *testing.T) {
	ts := mockArchiveServer(t, http.StatusNoContent)
	defer ts.Close()

	client := NewClient(ts.URL, "")
	_, err := client.ArchiveSecretVersion("secret", "myapp/config", 2)
	if err == nil {
		t.Fatal("expected permission denied error, got nil")
	}
}
