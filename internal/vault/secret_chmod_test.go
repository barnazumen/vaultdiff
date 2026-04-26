package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockChmodServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Vault-Token")
		if token != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/v1/secret/metadata/myapp/db" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"custom_metadata": map[string]string{
							"owner":       "alice",
							"read_roles":  "reader,auditor",
							"write_roles": "admin",
							"deny_roles":  "",
						},
					},
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPatch:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestGetSecretPermissions_Success(t *testing.T) {
	srv := mockChmodServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	perms, err := c.GetSecretPermissions("secret", "myapp/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perms.Owner != "alice" {
		t.Errorf("expected owner alice, got %s", perms.Owner)
	}
	if len(perms.ReadRoles) != 2 {
		t.Errorf("expected 2 read roles, got %d", len(perms.ReadRoles))
	}
	if perms.ReadRoles[0] != "reader" || perms.ReadRoles[1] != "auditor" {
		t.Errorf("unexpected read roles: %v", perms.ReadRoles)
	}
	if len(perms.WriteRoles) != 1 || perms.WriteRoles[0] != "admin" {
		t.Errorf("unexpected write roles: %v", perms.WriteRoles)
	}
	if len(perms.DenyRoles) != 0 {
		t.Errorf("expected no deny roles, got %v", perms.DenyRoles)
	}
}

func TestGetSecretPermissions_NotFound(t *testing.T) {
	srv := mockChmodServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	_, err := c.GetSecretPermissions("secret", "missing/path")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestGetSecretPermissions_InvalidToken(t *testing.T) {
	srv := mockChmodServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "bad-token")
	_, err := c.GetSecretPermissions("secret", "myapp/db")
	if err == nil {
		t.Fatal("expected permission denied error")
	}
}

func TestSetSecretPermissions_Success(t *testing.T) {
	srv := mockChmodServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	err := c.SetSecretPermissions("secret", "myapp/db", SecretPermissions{
		Owner:      "bob",
		ReadRoles:  []string{"reader"},
		WriteRoles: []string{"admin", "devops"},
		DenyRoles:  []string{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetSecretPermissions_InvalidToken(t *testing.T) {
	srv := mockChmodServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "bad-token")
	err := c.SetSecretPermissions("secret", "myapp/db", SecretPermissions{Owner: "bob"})
	if err == nil {
		t.Fatal("expected permission denied error")
	}
}
