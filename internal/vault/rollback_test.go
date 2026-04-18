package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockRollbackServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/data/myapp":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]string{"key": "old-value"},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/secret/data/myapp":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/metadata/myapp":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"versions": map[string]interface{}{
						"1": map[string]interface{}{},
						"2": map[string]interface{}{},
						"3": map[string]interface{}{},
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestRollbackToVersion_Success(t *testing.T) {
	srv := mockRollbackServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	result, err := client.RollbackToVersion("secret", "myapp", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.ToVersion != 3 {
		t.Errorf("expected ToVersion=3, got %d", result.ToVersion)
	}
	if result.Path != "myapp" {
		t.Errorf("expected Path=myapp, got %s", result.Path)
	}
}
