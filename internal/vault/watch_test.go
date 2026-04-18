package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func mockWatchServer(t *testing.T, calls *int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		var body string
		if *calls == 1 {
			body = `{"data":{"versions":{"1":{},"2":{}}}}`
		} else {
			body = `{"data":{"versions":{"1":{},"2":{},"3":{}}}}`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
}

func TestWatchSecret_DetectsNewVersion(t *testing.T) {
	calls := 0
	srv := mockWatchServer(t, &calls)
	defer srv.Close()

	c := &Client{Address: srv.URL, Token: "test-token", HTTPClient: srv.Client()}
	out := make(chan VersionChange, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	opts := WatchOptions{Interval: 100 * time.Millisecond}
	go WatchSecret(ctx, c, "secret/data/myapp", opts, out)

	select {
	case change := <-out:
		if change.FromVersion != 2 || change.ToVersion != 3 {
			t.Errorf("expected 2->3, got %d->%d", change.FromVersion, change.ToVersion)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for version change")
	}
}

func TestLatestVersion(t *testing.T) {
	versions := []int{1, 3, 2}
	if got := latestVersion(versions); got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestLatestVersion_Empty(t *testing.T) {
	if got := latestVersion([]int{}); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}
