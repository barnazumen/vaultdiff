package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockAnnotateServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch r.Method {
		case http.MethodPatch:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			if r.URL.Path == "/v1/secret/metadata/missing" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"custom_metadata": map[string]string{
						"owner": "alice",
						"team":  "platform",
					},
				},
			})
		}
	}))
}

func TestSetAnnotation_Success(t *testing.T) {
	srv := mockAnnotateServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	if err := c.SetAnnotation("secret", "myapp/config", "owner", "alice"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSetAnnotation_InvalidToken(t *testing.T) {
	srv := mockAnnotateServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "bad-token")
	if err := c.SetAnnotation("secret", "myapp/config", "owner", "alice"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestGetAnnotations_Success(t *testing.T) {
	srv := mockAnnotateServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	result, err := c.GetAnnotations("secret", "myapp/config")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Annotations["owner"] != "alice" {
		t.Errorf("expected owner=alice, got %q", result.Annotations["owner"])
	}
	if result.Annotations["team"] != "platform" {
		t.Errorf("expected team=platform, got %q", result.Annotations["team"])
	}
}

func TestGetAnnotations_NotFound(t *testing.T) {
	srv := mockAnnotateServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "test-token")
	_, err := c.GetAnnotations("secret", "missing")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestGetAnnotations_InvalidToken(t *testing.T) {
	srv := mockAnnotateServer(t)
	defer srv.Close()

	c := NewClient(srv.URL, "bad-token")
	_, err := c.GetAnnotations("secret", "myapp/config")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
