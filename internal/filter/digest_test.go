package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeDigestLease(id, path, sev string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  sev,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func TestDigest_ReturnsTopN(t *testing.T) {
	leases := []vault.SecretLease{
		makeDigestLease("id-1", "secret/a", "critical", 10*time.Minute),
		makeDigestLease("id-2", "secret/b", "warn", 20*time.Minute),
		makeDigestLease("id-3", "secret/c", "warn", 30*time.Minute),
		makeDigestLease("id-4", "secret/d", "ok", 60*time.Minute),
	}
	opts := DefaultDigestOptions()
	opts.TopN = 2
	opts.MinSev = "warn"

	entries := Digest(leases, opts)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].LeaseID != "id-1" {
		t.Errorf("expected first entry to be id-1, got %s", entries[0].LeaseID)
	}
}

func TestDigest_FiltersBelowMinSeverity(t *testing.T) {
	leases := []vault.SecretLease{
		makeDigestLease("id-ok", "secret/ok", "ok", 5*time.Minute),
		makeDigestLease("id-warn", "secret/warn", "warn", 15*time.Minute),
	}
	opts := DefaultDigestOptions()
	opts.MinSev = "warn"

	entries := Digest(leases, opts)
	for _, e := range entries {
		if e.Severity == "ok" {
			t.Errorf("ok severity should be filtered out")
		}
	}
}

func TestDigest_Empty(t *testing.T) {
	entries := Digest(nil, DefaultDigestOptions())
	if len(entries) != 0 {
		t.Errorf("expected empty digest for nil input")
	}
}

func TestDigest_SortedByExpiry(t *testing.T) {
	leases := []vault.SecretLease{
		makeDigestLease("id-far", "secret/far", "critical", 60*time.Minute),
		makeDigestLease("id-near", "secret/near", "critical", 5*time.Minute),
	}
	opts := DefaultDigestOptions()
	opts.MinSev = "warn"

	entries := Digest(leases, opts)
	if len(entries) < 2 {
		t.Fatal("expected at least 2 entries")
	}
	if entries[0].LeaseID != "id-near" {
		t.Errorf("expected nearest expiry first, got %s", entries[0].LeaseID)
	}
}

func TestPrintDigest_ContainsHeaders(t *testing.T) {
	entries := []DigestEntry{
		{LeaseID: "abc-123", Path: "secret/x", Severity: "critical", ExpiresIn: 5 * time.Minute},
	}
	var buf bytes.Buffer
	PrintDigest(entries, &buf)
	out := buf.String()
	if !strings.Contains(out, "LEASE ID") {
		t.Errorf("expected LEASE ID header in output")
	}
	if !strings.Contains(out, "EXPIRES IN") {
		t.Errorf("expected EXPIRES IN header in output")
	}
}

func TestPrintDigest_Empty(t *testing.T) {
	var buf bytes.Buffer
	PrintDigest(nil, &buf)
	if !strings.Contains(buf.String(), "No leases") {
		t.Errorf("expected empty message for nil entries")
	}
}
