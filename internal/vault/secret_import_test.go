package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func mockImportServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"version":1}}`))
	}))
}

func writeImportFile(t *testing.T, entries []ImportEntry) string {
	t.Helper()
	b, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	tmp := filepath.Join(t.TempDir(), "import.json")
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return tmp
}

func TestImportSecretsFromFile_Success(t *testing.T) {
	srv := mockImportServer(t)
	defer srv.Close()

	entries := []ImportEntry{
		{Path: "app/db", Data: map[string]string{"user": "admin", "pass": "secret"}},
		{Path: "app/api", Data: map[string]string{"key": "abc123"}},
	}
	file := writeImportFile(t, entries)

	results, err := ImportSecretsFromFile(srv.URL, "test-token", "secret", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for %s, got error: %s", r.Path, r.Error)
		}
	}
}

func TestImportSecretsFromFile_InvalidToken(t *testing.T) {
	srv := mockImportServer(t)
	defer srv.Close()

	entries := []ImportEntry{
		{Path: "app/db", Data: map[string]string{"user": "admin"}},
	}
	file := writeImportFile(t, entries)

	results, err := ImportSecretsFromFile(srv.URL, "bad-token", "secret", file)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(results) != 1 || results[0].Success {
		t.Errorf("expected failure for invalid token")
	}
}

func TestImportSecretsFromFile_InvalidFile(t *testing.T) {
	_, err := ImportSecretsFromFile("http://localhost", "tok", "secret", "/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
