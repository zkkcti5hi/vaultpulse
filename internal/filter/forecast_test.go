package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

func makeForecastLease(path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:        "lease/" + path,
		Path:           path,
		Severity:       severity,
		LeaseDuration:  int(ttl.Seconds()),
		Renewable:      false,
	}
}

func TestForecast_IncludesWithinWindow(t *testing.T) {
	leases := []vault.SecretLease{
		makeForecastLease("secret/app/db", "critical", 30*time.Minute),
		makeForecastLease("secret/app/api", "warn", 48*time.Hour),
		makeForecastLease("secret/app/cache", "ok", 200*time.Hour),
	}
	opts := ForecastOptions{Window: 72 * time.Hour, MinSeverity: "warn"}
	entries := Forecast(leases, opts)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestForecast_ExcludesExpired(t *testing.T) {
	leases := []vault.SecretLease{
		makeForecastLease("secret/old", "critical", -10*time.Minute),
		makeForecastLease("secret/soon", "critical", 1*time.Hour),
	}
	opts := ForecastOptions{Window: 72 * time.Hour, MinSeverity: "ok"}
	entries := Forecast(leases, opts)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Lease.Path != "secret/soon" {
		t.Errorf("unexpected path: %s", entries[0].Lease.Path)
	}
}

func TestForecast_SortedByExpiry(t *testing.T) {
	leases := []vault.SecretLease{
		makeForecastLease("secret/b", "warn", 10*time.Hour),
		makeForecastLease("secret/a", "warn", 2*time.Hour),
		makeForecastLease("secret/c", "warn", 5*time.Hour),
	}
	opts := ForecastOptions{Window: 72 * time.Hour, MinSeverity: "ok"}
	entries := Forecast(leases, opts)
	if len(entries) != 3 {
		t.Fatalf("expected 3, got %d", len(entries))
	}
	if entries[0].Lease.Path != "secret/a" {
		t.Errorf("expected secret/a first, got %s", entries[0].Lease.Path)
	}
}

func TestForecast_Empty(t *testing.T) {
	entries := Forecast(nil, DefaultForecastOptions())
	if len(entries) != 0 {
		t.Errorf("expected empty result")
	}
}

func TestPrintForecast_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeForecastLease("secret/app", "critical", 1*time.Hour),
	}
	opts := ForecastOptions{Window: 72 * time.Hour, MinSeverity: "ok"}
	entries := Forecast(leases, opts)
	var buf bytes.Buffer
	PrintForecast(entries, &buf)
	out := buf.String()
	for _, hdr := range []string{"PATH", "SEVERITY", "TTL", "EXPIRES AT"} {
		if !bytes.Contains([]byte(out), []byte(hdr)) {
			t.Errorf("missing header %q in output", hdr)
		}
	}
}

func TestPrintForecast_Empty(t *testing.T) {
	var buf bytes.Buffer
	PrintForecast(nil, &buf)
	if !bytes.Contains(buf.Bytes(), []byte("No leases")) {
		t.Errorf("expected empty message")
	}
}
