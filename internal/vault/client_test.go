package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockVaultServer(t *testing.T, version int, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"data": data,
				"metadata": map[string]interface{}{"version": version},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestReadSecretVersion_Success(t *testing.T) {
	expected := map[string]interface{}{"username": "admin", "password": "s3cr3t"}
	srv := mockVaultServer(t, 2, expected)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	got, err := client.ReadSecretVersion("secret", "myapp/config", 2)
	if err != nil {
		t.Fatalf("ReadSecretVersion: %v", err)
	}

	for k, v := range expected {
		if got[k] != v {
			t.Errorf("key %q: want %v, got %v", k, v, got[k])
		}
	}
}

func TestReadSecretVersion_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.ReadSecretVersion("secret", "missing/key", 1)
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}
