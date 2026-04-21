package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockKVCopyServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/v1/secret/data/src/key" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{"foo": "bar"},
					},
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		case http.MethodPost:
			if r.URL.Path == "/v1/secret/data/dst/key" {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestCopySecret_Success(t *testing.T) {
	srv := mockKVCopyServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	err := client.CopySecret(KVCopyOptions{
		SourceMount: "secret",
		DestMount:   "secret",
		SourcePath:  "src/key",
		DestPath:    "dst/key",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCopySecret_SourceNotFound(t *testing.T) {
	srv := mockKVCopyServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	err := client.CopySecret(KVCopyOptions{
		SourceMount: "secret",
		SourcePath:  "missing/key",
		DestPath:    "dst/key",
	})
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestCopySecret_InvalidToken(t *testing.T) {
	srv := mockKVCopyServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "bad-token")
	err := client.CopySecret(KVCopyOptions{
		SourceMount: "secret",
		SourcePath:  "src/key",
		DestPath:    "dst/key",
	})
	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
}
