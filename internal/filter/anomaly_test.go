package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/user/vaultpulse/internal/vault"
)

func makeAnomalyLease(path, severity string, expiresIn time.Duration, seenAgo time.Duration) vault.SecretLease {
	now := time.Now()
	l := vault.SecretLease{
		LeaseID:   path + "-id",
		Path:      path,
		Severity:  severity,
		ExpiresAt: now.Add(expiresIn),
	}
	if seenAgo >= 0 {
		l.SeenAt = now.Add(-seenAgo)
	}
	return l
}

func TestDetectAnomalies_ShortTTL(t *testing.T) {
	opts := DefaultAnomalyOptions()
	leases := []vault.SecretLease{
		makeAnomalyLease("secret/short", "critical", 2*time.Minute, -1),
		makeAnomalyLease("secret/normal", "ok", 2*time.Hour, -1),
	}
	results := DetectAnomalies(leases, opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(results))
	}
	if results[0].Lease.Path != "secret/short" {
		t.Errorf("expected secret/short, got %s", results[0].Lease.Path)
	}
}

func TestDetectAnomalies_RecentlySeen(t *testing.T) {
	opts := DefaultAnomalyOptions()
	leases := []vault.SecretLease{
		makeAnomalyLease("secret/new", "warn", 1*time.Hour, 30*time.Second),
		makeAnomalyLease("secret/old", "ok", 1*time.Hour, 30*time.Minute),
	}
	results := DetectAnomalies(leases, opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(results))
	}
	if results[0].Lease.Path != "secret/new" {
		t.Errorf("expected secret/new, got %s", results[0].Lease.Path)
	}
}

func TestDetectAnomalies_Empty(t *testing.T) {
	results := DetectAnomalies(nil, DefaultAnomalyOptions())
	if len(results) != 0 {
		t.Errorf("expected empty results")
	}
}

func TestDetectAnomalies_SortedByExpiry(t *testing.T) {
	opts := DefaultAnomalyOptions()
	leases := []vault.SecretLease{
		makeAnomalyLease("secret/b", "critical", 4*time.Minute, -1),
		makeAnomalyLease("secret/a", "critical", 1*time.Minute, -1),
	}
	results := DetectAnomalies(leases, opts)
	if len(results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Lease.Path != "secret/a" {
		t.Errorf("expected secret/a first, got %s", results[0].Lease.Path)
	}
}

func TestPrintAnomalies_NoResults(t *testing.T) {
	var buf bytes.Buffer
	PrintAnomalies(nil, &buf)
	if buf.Len() == 0 {
		t.Error("expected output for empty results")
	}
	if !bytes.Contains(buf.Bytes(), []byte("No anomalies")) {
		t.Errorf("expected 'No anomalies' message, got: %s", buf.String())
	}
}

func TestPrintAnomalies_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	results := []AnomalyResult{
		{Lease: makeAnomalyLease("secret/x", "critical", 2*time.Minute, -1), Reason: "short TTL"},
	}
	PrintAnomalies(results, &buf)
	for _, hdr := range []string{"PATH", "SEVERITY", "REASON"} {
		if !bytes.Contains(buf.Bytes(), []byte(hdr)) {
			t.Errorf("missing header %q in output", hdr)
		}
	}
}
