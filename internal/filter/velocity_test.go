package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeVelocityLease(path, severity string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "lease-" + path,
		Path:      path,
		TTL:       int(expiresIn.Seconds()),
		Severity:  severity,
		CreatedAt: time.Now(),
	}
}

func TestVelocity_CountsWithinWindow(t *testing.T) {
	leases := []vault.SecretLease{
		makeVelocityLease("secret/app/db", "critical", 2*time.Hour),
		makeVelocityLease("secret/app/api", "warn", 4*time.Hour),
		makeVelocityLease("secret/app/cache", "ok", 8*time.Hour),
	}
	opts := DefaultVelocityOptions()
	results := Velocity(leases, opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 prefix group, got %d", len(results))
	}
	if results[0].Count != 3 {
		t.Errorf("expected count 3, got %d", results[0].Count)
	}
	if results[0].Prefix != "secret/app" {
		t.Errorf("unexpected prefix %q", results[0].Prefix)
	}
}

func TestVelocity_ExcludesExpired(t *testing.T) {
	leases := []vault.SecretLease{
		makeVelocityLease("secret/old", "critical", -1*time.Hour),
		makeVelocityLease("secret/new", "warn", 2*time.Hour),
	}
	results := Velocity(leases, DefaultVelocityOptions())
	for _, r := range results {
		if r.Prefix == "secret/old" {
			t.Errorf("expired lease prefix should be excluded")
		}
	}
}

func TestVelocity_RatePerHour(t *testing.T) {
	leases := []vault.SecretLease{
		makeVelocityLease("secret/svc/a", "warn", 1*time.Hour),
		makeVelocityLease("secret/svc/b", "warn", 2*time.Hour),
	}
	opts := VelocityOptions{WindowSize: 2 * time.Hour, MinLeases: 1}
	results := Velocity(leases, opts)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	expected := 2.0 / 2.0
	if results[0].RatePerHour != expected {
		t.Errorf("expected rate %.4f, got %.4f", expected, results[0].RatePerHour)
	}
}

func TestVelocity_MinLeasesFilters(t *testing.T) {
	leases := []vault.SecretLease{
		makeVelocityLease("secret/rare/x", "ok", 1*time.Hour),
	}
	opts := VelocityOptions{WindowSize: 24 * time.Hour, MinLeases: 2}
	results := Velocity(leases, opts)
	if len(results) != 0 {
		t.Errorf("expected no results due to MinLeases filter, got %d", len(results))
	}
}

func TestVelocity_Empty(t *testing.T) {
	results := Velocity(nil, DefaultVelocityOptions())
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestPrintVelocity_ContainsHeaders(t *testing.T) {
	results := []VelocityResult{
		{Prefix: "secret/app", Count: 3, RatePerHour: 0.125, Severity: "warn"},
	}
	var buf bytes.Buffer
	PrintVelocity(results, &buf)
	out := buf.String()
	for _, hdr := range []string{"PREFIX", "COUNT", "RATE/HR", "SEVERITY"} {
		if !bytes.Contains([]byte(out), []byte(hdr)) {
			t.Errorf("output missing header %q", hdr)
		}
	}
}

func TestPrintVelocity_Empty(t *testing.T) {
	var buf bytes.Buffer
	PrintVelocity(nil, &buf)
	if !bytes.Contains(buf.Bytes(), []byte("no velocity data")) {
		t.Error("expected 'no velocity data' message")
	}
}
