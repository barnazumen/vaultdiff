package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func mockExportServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch r.URL.Path {
		case "/v1/secret/metadata/":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"alpha", "beta"},
				},
			})
		case "/v1/secret/data/alpha", "/v1/secret/data/beta":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data":    map[string]interface{}{"key": "value"},
					"metadata": map[string]interface{}{"version": float64(2)},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestExportSecrets_WritesValidFile(t *testing.T) {
	srv := mockExportServer(t)
	defer srv.Close()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "export.json")

	result, err := ExportSecrets(srv.URL, "test-token", "secret", outPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected 2 entries, got %d", result.Total)
	}

	if result.Mount != "secret" {
		t.Errorf("expected mount 'secret', got %s", result.Mount)
	}

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	defer f.Close()

	var parsed ExportResult
	if err := json.NewDecoder(f).Decode(&parsed); err != nil {
		t.Fatalf("invalid JSON in output file: %v", err)
	}

	if parsed.Total != 2 {
		t.Errorf("parsed total mismatch: got %d", parsed.Total)
	}
}

func TestExportSecrets_InvalidToken(t *testing.T) {
	srv := mockExportServer(t)
	defer srv.Close()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "export.json")

	_, err := ExportSecrets(srv.URL, "bad-token", "secret", outPath)
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}
