package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockMountsServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestListMounts_Success(t *testing.T) {
	payload := map[string]MountInfo{
		"secret/": {Type: "kv", Description: "key/value", Accessor: "kv_abc123"},
		"pki/":    {Type: "pki", Description: "PKI", Accessor: "pki_xyz"},
	}
	srv := mockMountsServer(t, http.StatusOK, payload)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	mounts, err := c.ListMounts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) != 2 {
		t.Errorf("expected 2 mounts, got %d", len(mounts))
	}
	if mounts["secret/"].Type != "kv" {
		t.Errorf("expected kv type, got %s", mounts["secret/"].Type)
	}
}

func TestListMounts_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	_, err := c.ListMounts()
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestListMounts_ServerError(t *testing.T) {
	srv := mockMountsServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	c := NewClient(srv.URL, "tok")
	_, err := c.ListMounts()
	if err == nil {
		t.Fatal("expected error on server error")
	}
}
