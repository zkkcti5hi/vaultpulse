package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeStaleLease(path, severity string, seenAgo time.Duration) vault.SecretLease {
	seenAt := time.Now().Add(-seenAgo).UTC().Format(time.RFC3339)
	return vault.SecretLease{
		LeaseID:  "lease/" + path,
		Path:     path,
		Severity: severity,
		Metadata: map[string]string{"seen_at": seenAt},
	}
}

func TestDetectStale_ReturnsOldLeases(t *testing.T) {
	leases := []vault.SecretLease{
		makeStaleLease("secret/old", "warn", 48*time.Hour),
		makeStaleLease("secret/fresh", "ok", 1*time.Hour),
		makeStaleLease("secret/ancient", "critical", 72*time.Hour),
	}
	opts := DefaultStaleOptions()
	got := DetectStale(leases, opts)
	if len(got) != 2 {
		t.Fatalf("expected 2 stale leases, got %d", len(got))
	}
}

func TestDetectStale_SortedOldestFirst(t *testing.T) {
	leases := []vault.SecretLease{
		makeStaleLease("secret/b", "warn", 30*time.Hour),
		makeStaleLease("secret/a", "warn", 60*time.Hour),
	}
	opts := DefaultStaleOptions()
	got := DetectStale(leases, opts)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0].Path != "secret/a" {
		t.Errorf("expected secret/a first, got %s", got[0].Path)
	}
}

func TestDetectStale_Empty(t *testing.T) {
	got := DetectStale(nil, DefaultStaleOptions())
	if got != nil {
		t.Errorf("expected nil for empty input")
	}
}

func TestDetectStale_NoSeenAt_Skipped(t *testing.T) {
	l := vault.SecretLease{LeaseID: "x", Path: "secret/x", Severity: "warn", Metadata: map[string]string{}}
	got := DetectStale([]vault.SecretLease{l}, DefaultStaleOptions())
	if len(got) != 0 {
		t.Errorf("expected lease without seen_at to be skipped")
	}
}

func TestDetectStale_MaxResults(t *testing.T) {
	var leases []vault.SecretLease
	for i := 0; i < 10; i++ {
		leases = append(leases, makeStaleLease("secret/p", "warn", 48*time.Hour))
	}
	opts := DefaultStaleOptions()
	opts.MaxResults = 3
	got := DetectStale(leases, opts)
	if len(got) != 3 {
		t.Errorf("expected 3 results, got %d", len(got))
	}
}

func TestPrintStale_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeStaleLease("secret/old", "warn", 48*time.Hour),
	}
	var buf bytes.Buffer
	PrintStale(leases, DefaultStaleOptions(), &buf)
	out := buf.String()
	if !strings.Contains(out, "PATH") || !strings.Contains(out, "SEVERITY") || !strings.Contains(out, "LAST SEEN") {
		t.Errorf("output missing headers: %s", out)
	}
}

func TestPrintStale_EmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	PrintStale(nil, DefaultStaleOptions(), &buf)
	if !strings.Contains(buf.String(), "No stale leases") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
