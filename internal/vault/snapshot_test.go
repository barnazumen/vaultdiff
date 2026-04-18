package vault

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockSnapshotServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/myapp":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"versions": map[string]interface{}{
						"1": map[string]interface{}{},
						"2": map[string]interface{}{},
					},
				},
			})
		case "/v1/secret/data/myapp":
			v := r.URL.Query().Get("version")
			data := map[string]interface{}{"key": "val" + v}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestExportSnapshot_WritesAllVersions(t *testing.T) {
	srv := mockSnapshotServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	var buf bytes.Buffer
	if err := client.ExportSnapshot("myapp", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var snap SecretSnapshot
	if err := json.NewDecoder(&buf).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if snap.Path != "myapp" {
		t.Errorf("expected path myapp, got %s", snap.Path)
	}
	if len(snap.Versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(snap.Versions))
	}
}

func TestLoadSnapshot_InvalidFile(t *testing.T) {
	_, err := LoadSnapshot("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
