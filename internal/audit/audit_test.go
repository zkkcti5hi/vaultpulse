package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/you/vaultpulse/internal/audit"
	"github.com/you/vaultpulse/internal/vault"
)

func makeLease(id, path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestLogger_WritesJSONLines(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)
	leases := []vault.SecretLease{
		makeLease("lease/1", "secret/db", "critical", 5*time.Minute),
		makeLease("lease/2", "secret/api", "warning", 20*time.Minute),
	}
	if err := l.Log(leases); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	var entry audit.Entry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.LeaseID != "lease/1" {
		t.Errorf("expected lease/1, got %s", entry.LeaseID)
	}
	if entry.Severity != "critical" {
		t.Errorf("expected critical, got %s", entry.Severity)
	}
}

func TestLogger_EmptyLeasesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)
	if err := l.Log(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestLogger_DefaultsToStderr(t *testing.T) {
	l := audit.NewLogger(nil)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}
