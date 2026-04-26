package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeRetentionLease(id, path string, ttlSeconds int, seenAt time.Time) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		TTL:      ttlSeconds,
		SeenAt:   seenAt,
		Severity: "ok",
	}
}

func TestApplyRetention_NoViolations(t *testing.T) {
	policy := RetentionPolicy{
		MaxAge: 30 * 24 * time.Hour,
		MaxTTL: 90 * 24 * time.Hour,
	}
	leases := []vault.SecretLease{
		makeRetentionLease("lease-1", "secret/a", 3600, time.Now().Add(-1*time.Hour)),
	}
	results := ApplyRetention(leases, policy)
	if len(results) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(results))
	}
}

func TestApplyRetention_AgeViolation(t *testing.T) {
	policy := RetentionPolicy{
		MaxAge: 7 * 24 * time.Hour,
		MaxTTL: 0,
	}
	oldSeen := time.Now().Add(-10 * 24 * time.Hour)
	leases := []vault.SecretLease{
		makeRetentionLease("lease-old", "secret/old", 3600, oldSeen),
		makeRetentionLease("lease-new", "secret/new", 3600, time.Now().Add(-1*time.Hour)),
	}
	results := ApplyRetention(leases, policy)
	if len(results) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(results))
	}
	if results[0].Lease.LeaseID != "lease-old" {
		t.Errorf("expected lease-old, got %s", results[0].Lease.LeaseID)
	}
}

func TestApplyRetention_TTLViolation(t *testing.T) {
	policy := RetentionPolicy{
		MaxAge: 0,
		MaxTTL: 24 * time.Hour,
	}
	// TTL of 2 days in seconds
	leases := []vault.SecretLease{
		makeRetentionLease("lease-longtll", "secret/long", 2*86400, time.Now()),
		makeRetentionLease("lease-short", "secret/short", 3600, time.Now()),
	}
	results := ApplyRetention(leases, policy)
	if len(results) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(results))
	}
	if results[0].Lease.LeaseID != "lease-longtll" {
		t.Errorf("expected lease-longtll, got %s", results[0].Lease.LeaseID)
	}
}

func TestApplyRetention_Empty(t *testing.T) {
	policy := DefaultRetentionOptions()
	results := ApplyRetention([]vault.SecretLease{}, policy)
	if len(results) != 0 {
		t.Fatalf("expected empty results, got %d", len(results))
	}
}

func TestPrintRetention_ContainsHeaders(t *testing.T) {
	results := []RetentionResult{
		{
			Lease:  makeRetentionLease("l1", "secret/x", 3600, time.Now()),
			Reason: "age 10d exceeds max 7d",
			Age:    10 * 24 * time.Hour,
		},
	}
	var buf bytes.Buffer
	PrintRetention(results, &buf)
	out := buf.String()
	for _, hdr := range []string{"LEASE ID", "PATH", "REASON"} {
		if !bytes.Contains([]byte(out), []byte(hdr)) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestPrintRetention_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	PrintRetention(nil, &buf)
	if !bytes.Contains(buf.Bytes(), []byte("No retention violations")) {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
