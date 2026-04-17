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

func TestLogger_EntryFields(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)
	lease := vault.SecretLease{
		LeaseID:   "lease/test/xyz",
		Path:      "secret/data/myapp",
		Severity:  "warning",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := l.Log([]vault.SecretLease{lease}); err != nil {
		t.Fatalf("Log returned error: %v", err)
	}
	line := strings.TrimSpace(buf.String())
	var entry audit.Entry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("could not parse entry: %v", err)
	}
	if entry.Path != "secret/data/myapp" {
		t.Errorf("path mismatch: %s", entry.Path)
	}
	if entry.Message == "" {
		t.Error("expected non-empty message")
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if entry.TTL == "" {
		t.Error("expected non-empty TTL")
	}
}
