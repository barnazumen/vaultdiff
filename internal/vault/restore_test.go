package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockRestoreServer(t *testing.T, written *[]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if data, ok := body["data"].(map[string]interface{}); ok {
			*written = append(*written, data)
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func TestRestoreFromSnapshot_Success(t *testing.T) {
	var written []map[string]interface{}
	srv := mockRestoreServer(t, &written)
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")

	snapshot := []SecretSnapshot{
		{Version: 1, Data: map[string]interface{}{"key": "val1"}},
		{Version: 2, Data: map[string]interface{}{"key": "val2"}},
	}

	err := client.RestoreFromSnapshot("secret", "myapp/config", snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(written) != 2 {
		t.Fatalf("expected 2 writes, got %d", len(written))
	}
	if written[0]["key"] != "val1" {
		t.Errorf("expected val1, got %v", written[0]["key"])
	}
}

func TestRestoreFromSnapshot_InvalidToken(t *testing.T) {
	var written []map[string]interface{}
	srv := mockRestoreServer(t, &written)
	defer srv.Close()

	client := NewClient(srv.URL, "bad-token")
	snapshot := []SecretSnapshot{
		{Version: 1, Data: map[string]interface{}{"key": "val"}},
	}

	err := client.RestoreFromSnapshot("secret", "myapp/config", snapshot)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
