package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeRollupLease(path, severity string, tags []string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "lease/" + path,
		Path:      path,
		Severity:  severity,
		Tags:      tags,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func TestRollup_BySeverity(t *testing.T) {
	leases := []vault.SecretLease{
		makeRollupLease("secret/a", "critical", nil, time.Minute),
		makeRollupLease("secret/b", "critical", nil, 2*time.Minute),
		makeRollupLease("secret/c", "warn", nil, 10*time.Minute),
		makeRollupLease("secret/d", "ok", nil, time.Hour),
	}
	entries := Rollup(leases, RollupOptions{GroupBy: "severity"})
	keys := map[string]RollupEntry{}
	for _, e := range entries {
		keys[e.Key] = e
	}
	if keys["critical"].Count != 2 {
		t.Errorf("expected 2 critical, got %d", keys["critical"].Count)
	}
	if keys["warn"].Count != 1 {
		t.Errorf("expected 1 warn, got %d", keys["warn"].Count)
	}
	if keys["ok"].Count != 1 {
		t.Errorf("expected 1 ok, got %d", keys["ok"].Count)
	}
}

func TestRollup_ByPath(t *testing.T) {
	leases := []vault.SecretLease{
		makeRollupLease("secret/db/pass", "critical", nil, time.Minute),
		makeRollupLease("secret/db/user", "warn", nil, 5*time.Minute),
		makeRollupLease("secret/api/key", "ok", nil, time.Hour),
	}
	entries := Rollup(leases, RollupOptions{GroupBy: "path"})
	keys := map[string]RollupEntry{}
	for _, e := range entries {
		keys[e.Key] = e
	}
	if keys["secret"].Count != 3 {
		t.Errorf("expected 3 under 'secret', got %d", keys["secret"].Count)
	}
}

func TestRollup_ByTag(t *testing.T) {
	leases := []vault.SecretLease{
		makeRollupLease("secret/a", "critical", []string{"team:ops"}, time.Minute),
		makeRollupLease("secret/b", "warn", []string{"team:ops", "env:prod"}, 5*time.Minute),
		makeRollupLease("secret/c", "ok", nil, time.Hour),
	}
	entries := Rollup(leases, RollupOptions{GroupBy: "tag"})
	keys := map[string]RollupEntry{}
	for _, e := range entries {
		keys[e.Key] = e
	}
	if keys["team:ops"].Count != 2 {
		t.Errorf("expected 2 for team:ops, got %d", keys["team:ops"].Count)
	}
	if keys["(untagged)"].Count != 1 {
		t.Errorf("expected 1 untagged, got %d", keys["(untagged)"].Count)
	}
}

func TestRollup_Empty(t *testing.T) {
	entries := Rollup(nil, DefaultRollupOptions())
	if len(entries) != 0 {
		t.Errorf("expected empty result, got %d entries", len(entries))
	}
}

func TestPrintRollup_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeRollupLease("secret/x", "critical", nil, time.Minute),
	}
	entries := Rollup(leases, DefaultRollupOptions())
	var buf bytes.Buffer
	PrintRollup(entries, &buf)
	out := buf.String()
	for _, hdr := range []string{"KEY", "COUNT", "CRITICAL", "WARN", "EARLIEST_EXPIRY"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}
