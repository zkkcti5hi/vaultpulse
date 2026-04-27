package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeDriftLease(path string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "lease/" + path,
		Path:      path,
		ExpiresAt: time.Now().Add(ttl),
		Severity:  "ok",
	}
}

func TestDetectDrift_Empty(t *testing.T) {
	results := DetectDrift(nil, DefaultDriftOptions())
	if len(results) != 0 {
		t.Fatalf("expected empty results, got %d", len(results))
	}
}

func TestDetectDrift_ExactBaseline(t *testing.T) {
	opts := DefaultDriftOptions()
	lease := makeDriftLease("secret/exact", opts.BaselineTTL)
	results := DetectDrift([]vault.SecretLease{lease}, opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Exceeds {
		t.Error("lease at exact baseline should not exceed tolerance")
	}
	if results[0].DriftPct > 0.001 || results[0].DriftPct < -0.001 {
		t.Errorf("drift should be ~0, got %f", results[0].DriftPct)
	}
}

func TestDetectDrift_ShortTTL_Exceeds(t *testing.T) {
	opts := DefaultDriftOptions() // baseline 24h, tolerance 20%
	// TTL of 1h is way below baseline
	lease := makeDriftLease("secret/short", 1*time.Hour)
	results := DetectDrift([]vault.SecretLease{lease}, opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if !results[0].Exceeds {
		t.Error("short TTL should exceed tolerance")
	}
	if results[0].DriftPct >= 0 {
		t.Error("drift should be negative for short TTL")
	}
}

func TestDetectDrift_LongTTL_Exceeds(t *testing.T) {
	opts := DefaultDriftOptions()
	lease := makeDriftLease("secret/long", 72*time.Hour)
	results := DetectDrift([]vault.SecretLease{lease}, opts)
	if !results[0].Exceeds {
		t.Error("long TTL should exceed tolerance")
	}
	if results[0].DriftPct <= 0 {
		t.Error("drift should be positive for long TTL")
	}
}

func TestDetectDrift_SortedByAbsDriftDesc(t *testing.T) {
	opts := DefaultDriftOptions()
	leases := []vault.SecretLease{
		makeDriftLease("secret/slight", 22*time.Hour), // ~8% drift
		makeDriftLease("secret/huge", 1*time.Hour),   // ~96% drift
		makeDriftLease("secret/mid", 18*time.Hour),   // ~25% drift
	}
	results := DetectDrift(leases, opts)
	if results[0].Lease.Path != "secret/huge" {
		t.Errorf("expected highest drift first, got %s", results[0].Lease.Path)
	}
}

func TestPrintDrift_ContainsHeaders(t *testing.T) {
	opts := DefaultDriftOptions()
	leases := []vault.SecretLease{makeDriftLease("secret/a", 12*time.Hour)}
	results := DetectDrift(leases, opts)
	var buf bytes.Buffer
	PrintDrift(results, &buf)
	out := buf.String()
	if !contains(out, "LEASE PATH") {
		t.Error("output should contain LEASE PATH header")
	}
	if !contains(out, "DRIFT %") {
		t.Error("output should contain DRIFT % header")
	}
}
