package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockCloneServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/src/mykey":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"username": "admin", "password": "s3cr3t"},
					"metadata": map[string]interface{}{"version": 3},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/missing/key":
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/secret/data/dst/mykey":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 1}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestCloneSecret_Success(t *testing.T) {
	srv := mockCloneServer(t)
	defer srv.Close()

	result, err := CloneSecret(srv.URL, "test-token", CloneSecretOptions{
		SourcePath: "src/mykey",
		DestPath:   "dst/mykey",
		Mount:      "secret",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.SourcePath != "src/mykey" {
		t.Errorf("expected source path src/mykey, got %s", result.SourcePath)
	}
	if result.DestPath != "dst/mykey" {
		t.Errorf("expected dest path dst/mykey, got %s", result.DestPath)
	}
	if len(result.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result.Keys))
	}
}

func TestCloneSecret_SourceNotFound(t *testing.T) {
	srv := mockCloneServer(t)
	defer srv.Close()

	_, err := CloneSecret(srv.URL, "test-token", CloneSecretOptions{
		SourcePath: "missing/key",
		DestPath:   "dst/key",
		Mount:      "secret",
	})
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestCloneSecret_InvalidToken(t *testing.T) {
	srv := mockCloneServer(t)
	defer srv.Close()

	_, err := CloneSecret(srv.URL, "bad-token", CloneSecretOptions{
		SourcePath: "src/mykey",
		DestPath:   "dst/mykey",
		Mount:      "secret",
	})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
