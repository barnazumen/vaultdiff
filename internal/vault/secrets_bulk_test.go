package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

func mockBulkServer(t *testing.T) *httptest.Server {
	t.Helper()
	secrets := map[string]map[string]interface{}{
		"alpha": {"key": "val-a"},
		"beta":  {"key": "val-b"},
		"gamma": {"key": "val-g"},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// extract last path segment as secret name
		path := r.URL.Path
		name := path[len("/v1/secret/data/"):]
		data, ok := secrets[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		})
	}))
}

func TestReadSecretsBulk_AllFound(t *testing.T) {
	srv := mockBulkServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	paths := []string{"alpha", "beta", "gamma"}
	results := c.ReadSecretsBulk("secret", paths, 0)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Error)
		}
		if r.Data == nil {
			t.Errorf("expected data for %s", r.Path)
		}
	}
}

func TestReadSecretsBulk_PartialNotFound(t *testing.T) {
	srv := mockBulkServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	paths := []string{"alpha", "missing"}
	results := c.ReadSecretsBulk("secret", paths, 0)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Path < results[j].Path })
	if results[0].Error != nil {
		t.Errorf("expected no error for alpha, got %v", results[0].Error)
	}
	if results[1].Error == nil {
		t.Errorf("expected error for missing, got nil")
	}
}

func TestReadSecretsBulk_InvalidToken(t *testing.T) {
	srv := mockBulkServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "bad-token")
	results := c.ReadSecretsBulk("secret", []string{"alpha"}, 0)

	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if results[0].Error == nil {
		t.Error("expected error for invalid token")
	}
}

func TestReadSecretsBulk_Empty(t *testing.T) {
	srv := mockBulkServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	results := c.ReadSecretsBulk("secret", []string{}, 0)

	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}
