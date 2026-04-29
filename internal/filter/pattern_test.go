package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makePatternLease(path, severity, id string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		Severity: severity,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
}

func TestMatchPattern_Exact(t *testing.T) {
	if !MatchPattern("secret/db/prod", "secret/db/prod") {
		t.Error("expected exact match")
	}
	if MatchPattern("secret/db/prod", "secret/db/staging") {
		t.Error("expected no match")
	}
}

func TestMatchPattern_Wildcard(t *testing.T) {
	if !MatchPattern("secret/db/*", "secret/db/prod") {
		t.Error("expected wildcard suffix match")
	}
	if MatchPattern("secret/db/*", "secret/kv/prod") {
		t.Error("expected no match for different prefix")
	}
}

func TestMatchPattern_StarOnly(t *testing.T) {
	if !MatchPattern("*", "anything/goes") {
		t.Error("expected star to match everything")
	}
}

func TestFilterByPattern_NoPatterns(t *testing.T) {
	leases := []vault.SecretLease{
		makePatternLease("secret/a", "ok", "id-1"),
		makePatternLease("secret/b", "warn", "id-2"),
	}
	opts := DefaultPatternOptions()
	result := FilterByPattern(leases, opts)
	if len(result) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(result))
	}
}

func TestFilterByPattern_MatchesSingle(t *testing.T) {
	leases := []vault.SecretLease{
		makePatternLease("secret/db/prod", "critical", "id-1"),
		makePatternLease("secret/kv/prod", "ok", "id-2"),
	}
	opts := PatternOptions{Patterns: []string{"secret/db/*"}}
	result := FilterByPattern(leases, opts)
	if len(result) != 1 {
		t.Fatalf("expected 1 lease, got %d", len(result))
	}
	if result[0].LeaseID != "id-1" {
		t.Errorf("unexpected lease: %s", result[0].LeaseID)
	}
}

func TestFilterByPattern_Invert(t *testing.T) {
	leases := []vault.SecretLease{
		makePatternLease("secret/db/prod", "critical", "id-1"),
		makePatternLease("secret/kv/prod", "ok", "id-2"),
	}
	opts := PatternOptions{Patterns: []string{"secret/db/*"}, Invert: true}
	result := FilterByPattern(leases, opts)
	if len(result) != 1 {
		t.Fatalf("expected 1 lease, got %d", len(result))
	}
	if result[0].LeaseID != "id-2" {
		t.Errorf("unexpected lease: %s", result[0].LeaseID)
	}
}

func TestPrintPattern_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makePatternLease("secret/db/prod", "critical", "id-1"),
	}
	opts := PatternOptions{Patterns: []string{"secret/db/*"}}
	var buf bytes.Buffer
	PrintPattern(leases, opts, &buf)
	out := buf.String()
	for _, want := range []string{"PATH", "SEVERITY", "LEASE ID", "secret/db/prod", "1 lease(s)", "secret/db/*"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("output missing %q", want)
		}
	}
}
