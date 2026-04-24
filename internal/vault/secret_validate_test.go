package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockValidateServer(t *testing.T, statusCode int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(statusCode)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestValidateSecretKeys_AllPresent(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"data": map[string]interface{}{"username": "admin", "password": "s3cr3t"},
			"metadata": map[string]interface{}{"version": 2},
		},
	}
	srv := mockValidateServer(t, http.StatusOK, payload)
	defer srv.Close()

	res, err := ValidateSecretKeys(srv.URL, "tok", "secret", "myapp/db", 2, []string{"username", "password"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Valid {
		t.Errorf("expected valid, got missing=%v extra=%v", res.Missing, res.Extra)
	}
}

func TestValidateSecretKeys_MissingKey(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"data":     map[string]interface{}{"username": "admin"},
			"metadata": map[string]interface{}{"version": 1},
		},
	}
	srv := mockValidateServer(t, http.StatusOK, payload)
	defer srv.Close()

	res, err := ValidateSecretKeys(srv.URL, "tok", "secret", "myapp/db", 1, []string{"username", "password"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Error("expected invalid result")
	}
	if len(res.Missing) != 1 || res.Missing[0] != "password" {
		t.Errorf("expected missing=[password], got %v", res.Missing)
	}
}

func TestValidateSecretKeys_StrictMode_ExtraKey(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"data":     map[string]interface{}{"username": "admin", "extra_key": "oops"},
			"metadata": map[string]interface{}{"version": 1},
		},
	}
	srv := mockValidateServer(t, http.StatusOK, payload)
	defer srv.Close()

	res, err := ValidateSecretKeys(srv.URL, "tok", "secret", "myapp/db", 1, []string{"username"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Error("expected invalid in strict mode")
	}
	if len(res.Extra) == 0 {
		t.Error("expected extra keys reported")
	}
}

func TestValidateSecretKeys_NotFound(t *testing.T) {
	srv := mockValidateServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	_, err := ValidateSecretKeys(srv.URL, "tok", "secret", "missing/path", 1, []string{"key"}, false)
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestValidateSecretKeys_InvalidToken(t *testing.T) {
	srv := mockValidateServer(t, http.StatusForbidden, nil)
	defer srv.Close()

	_, err := ValidateSecretKeys(srv.URL, "", "secret", "myapp/db", 1, []string{"key"}, false)
	if err == nil {
		t.Fatal("expected permission denied error")
	}
}
