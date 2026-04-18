package vault_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultdiff/internal/diff"
	"github.com/your-org/vaultdiff/internal/vault"
)

func mockDiffServer(t *testing.T, dataA, dataB map[string]interface{}) *httptest.Server {
	t.Helper()
	call := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		var data map[string]interface{}
		if call == 1 {
			data = dataA
		} else {
			data = dataB
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		})
	}))
}

func TestDiffVersions_DetectsChanges(t *testing.T) {
	dataA := map[string]interface{}{"key1": "old", "key2": "same"}
	dataB := map[string]interface{}{"key1": "new", "key2": "same", "key3": "added"}

	srv := mockDiffServer(t, dataA, dataB)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	changes, err := vault.DiffVersions(client, "secret", "myapp/config", 1, 2)
	if err != nil {
		t.Fatalf("DiffVersions: %v", err)
	}

	counts := map[diff.ChangeType]int{}
	for _, c := range changes {
		counts[c.Type]++
	}

	if counts[diff.Modified] != 1 {
		t.Errorf("expected 1 modified, got %d", counts[diff.Modified])
	}
	if counts[diff.Added] != 1 {
		t.Errorf("expected 1 added, got %d", counts[diff.Added])
	}
	if counts[diff.Unchanged] != 1 {
		t.Errorf("expected 1 unchanged, got %d", counts[diff.Unchanged])
	}
	fmt.Println("DiffVersions changes:", changes)
}
