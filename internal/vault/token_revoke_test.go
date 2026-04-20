package vault_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func mockRevokeServer(t *testing.T, selfRevoke bool, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if selfRevoke && r.URL.Path != "/v1/auth/token/revoke-self" {
			t.Errorf("expected revoke-self path, got %s", r.URL.Path)
		}
		if !selfRevoke && r.URL.Path != "/v1/auth/token/revoke" {
			t.Errorf("expected revoke path, got %s", r.URL.Path)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestRevokeToken_SelfRevoke_Success(t *testing.T) {
	srv := mockRevokeServer(t, true, http.StatusNoContent)
	defer srv.Close()

	c := vault.NewClient(srv.URL, "test-token")
	err := c.RevokeToken("", true)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRevokeToken_ExplicitToken_Success(t *testing.T) {
	srv := mockRevokeServer(t, false, http.StatusNoContent)
	defer srv.Close()

	c := vault.NewClient(srv.URL, "test-token")
	err := c.RevokeToken("s.sometoken", false)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRevokeToken_InvalidToken(t *testing.T) {
	srv := mockRevokeServer(t, false, http.StatusForbidden)
	defer srv.Close()

	c := vault.NewClient(srv.URL, "bad-token")
	err := c.RevokeToken("s.sometoken", false)
	if err == nil {
		t.Fatal("expected error for forbidden, got nil")
	}
}

func TestRevokeToken_ServerError(t *testing.T) {
	srv := mockRevokeServer(t, false, http.StatusInternalServerError)
	defer srv.Close()

	c := vault.NewClient(srv.URL, "test-token")
	err := c.RevokeToken("s.sometoken", false)
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}
