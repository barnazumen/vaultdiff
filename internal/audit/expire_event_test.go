package audit

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func TestLogExpireEvent_NotExpired(t *testing.T) {
	var buf bytes.Buffer
	expiry := &vault.SecretExpiry{
		Path:        "myapp/db",
		Version:     3,
		CreatedTime: time.Now().Add(-10 * 24 * time.Hour),
		DaysOld:     10,
		Expired:     false,
	}

	if err := LogExpireEvent(&buf, expiry, 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry ExpireEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if entry.Event != "secret_expiry_check" {
		t.Errorf("expected event=secret_expiry_check, got %s", entry.Event)
	}
	if entry.Path != "myapp/db" {
		t.Errorf("expected path=myapp/db, got %s", entry.Path)
	}
	if entry.Status != "ok" {
		t.Errorf("expected status=ok, got %s", entry.Status)
	}
	if entry.Expired {
		t.Error("expected expired=false")
	}
}

func TestLogExpireEvent_Expired(t *testing.T) {
	var buf bytes.Buffer
	expiry := &vault.SecretExpiry{
		Path:        "myapp/creds",
		Version:     1,
		CreatedTime: time.Now().Add(-90 * 24 * time.Hour),
		DaysOld:     90,
		Expired:     true,
	}

	if err := LogExpireEvent(&buf, expiry, 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry ExpireEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if entry.Status != "expired" {
		t.Errorf("expected status=expired, got %s", entry.Status)
	}
	if !entry.Expired {
		t.Error("expected expired=true")
	}
	if entry.MaxAgeDays != 30 {
		t.Errorf("expected max_age_days=30, got %d", entry.MaxAgeDays)
	}
}
