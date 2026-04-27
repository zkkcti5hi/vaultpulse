package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/vaultpulse/internal/vault"
)

func makeWindowLease(path string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "lease/" + path,
		Path:      path,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func TestRollingWindow_BucketsCount(t *testing.T) {
	opts := DefaultWindowOptions()
	opts.NumWindows = 4
	buckets := RollingWindow(nil, opts)
	if len(buckets) != 4 {
		t.Fatalf("expected 4 buckets, got %d", len(buckets))
	}
}

func TestRollingWindow_PlacesLeaseInCorrectBucket(t *testing.T) {
	opts := DefaultWindowOptions()
	opts.WindowSize = time.Hour
	opts.NumWindows = 3

	leases := []vault.SecretLease{
		makeWindowLease("secret/a", 30*time.Minute),  // bucket 0: 0-1h
		makeWindowLease("secret/b", 90*time.Minute),  // bucket 1: 1-2h
		makeWindowLease("secret/c", 150*time.Minute), // bucket 2: 2-3h
	}

	buckets := RollingWindow(leases, opts)
	if len(buckets[0].Leases) != 1 {
		t.Errorf("bucket 0: expected 1 lease, got %d", len(buckets[0].Leases))
	}
	if len(buckets[1].Leases) != 1 {
		t.Errorf("bucket 1: expected 1 lease, got %d", len(buckets[1].Leases))
	}
	if len(buckets[2].Leases) != 1 {
		t.Errorf("bucket 2: expected 1 lease, got %d", len(buckets[2].Leases))
	}
}

func TestRollingWindow_ExcludesExpired(t *testing.T) {
	opts := DefaultWindowOptions()
	leases := []vault.SecretLease{
		makeWindowLease("secret/expired", -5*time.Minute),
		makeWindowLease("secret/valid", 30*time.Minute),
	}
	buckets := RollingWindow(leases, opts)
	total := 0
	for _, b := range buckets {
		total += len(b.Leases)
	}
	if total != 1 {
		t.Errorf("expected 1 non-expired lease across all buckets, got %d", total)
	}
}

func TestRollingWindow_Empty(t *testing.T) {
	opts := DefaultWindowOptions()
	buckets := RollingWindow(nil, opts)
	for _, b := range buckets {
		if len(b.Leases) != 0 {
			t.Errorf("expected empty bucket %s, got %d leases", b.Label, len(b.Leases))
		}
	}
}

func TestPrintWindows_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultWindowOptions()
	opts.Out = &buf
	buckets := RollingWindow(nil, opts)
	PrintWindows(buckets, opts)
	out := buf.String()
	if !strings.Contains(out, "WINDOW") {
		t.Error("expected WINDOW header")
	}
	if !strings.Contains(out, "COUNT") {
		t.Error("expected COUNT header")
	}
}

func TestPrintWindows_ShowsLeaseCount(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultWindowOptions()
	opts.WindowSize = time.Hour
	opts.NumWindows = 2
	opts.Out = &buf

	leases := []vault.SecretLease{
		makeWindowLease("secret/x", 20*time.Minute),
		makeWindowLease("secret/y", 25*time.Minute),
	}
	buckets := RollingWindow(leases, opts)
	PrintWindows(buckets, opts)
	out := buf.String()
	if !strings.Contains(out, "2") {
		t.Errorf("expected count 2 in output: %s", out)
	}
}
