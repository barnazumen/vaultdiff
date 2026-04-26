package vault_test

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func mockDiffExportServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch {
		case strings.Contains(r.URL.Path, "/v1/secret/data/myapp") && strings.Contains(r.URL.RawQuery, "version=1"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"username": "alice",
						"password": "old-pass",
					},
				},
			})
		case strings.Contains(r.URL.Path, "/v1/secret/data/myapp") && strings.Contains(r.URL.RawQuery, "version=2"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{
						"username": "alice",
						"password": "new-pass",
						"email":    "alice@example.com",
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestExportDiffToFile_JSON(t *testing.T) {
	srv := mockDiffExportServer(t)
	defer srv.Close()

	client := vault.NewClient(srv.URL, "test-token")
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "diff.json")

	err := vault.ExportDiffToFile(client, "secret", "myapp", 1, 2, outPath, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	changes, ok := result["changes"].([]interface{})
	if !ok || len(changes) == 0 {
		t.Fatalf("expected changes in output, got: %v", result)
	}
}

func TestExportDiffToFile_CSV(t *testing.T) {
	srv := mockDiffExportServer(t)
	defer srv.Close()

	client := vault.NewClient(srv.URL, "test-token")
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "diff.csv")

	err := vault.ExportDiffToFile(client, "secret", "myapp", 1, 2, outPath, "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatalf("failed to open output file: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	if len(records) < 2 {
		t.Fatalf("expected header + at least one data row, got %d rows", len(records))
	}

	header := records[0]
	expectedCols := []string{"key", "change_type", "old_value", "new_value"}
	for i, col := range expectedCols {
		if i >= len(header) || header[i] != col {
			t.Errorf("expected column %q at index %d, got %q", col, i, header[i])
		}
	}
}

func TestExportDiffToFile_InvalidToken(t *testing.T) {
	srv := mockDiffExportServer(t)
	defer srv.Close()

	client := vault.NewClient(srv.URL, "bad-token")
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "diff.json")

	err := vault.ExportDiffToFile(client, "secret", "myapp", 1, 2, outPath, "json")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestExportDiffToFile_UnsupportedFormat(t *testing.T) {
	srv := mockDiffExportServer(t)
	defer srv.Close()

	client := vault.NewClient(srv.URL, "test-token")
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "diff.xml")

	err := vault.ExportDiffToFile(client, "secret", "myapp", 1, 2, outPath, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error message, got: %v", err)
	}
}
