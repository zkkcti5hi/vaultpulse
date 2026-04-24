package filter_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeIntForecastLease(path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:       "lease/" + path,
		Path:          path,
		Severity:      severity,
		LeaseDuration: int(ttl.Seconds()),
	}
}

func TestForecast_ThenSort_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntForecastLease("secret/z", "warn", 50*time.Hour),
		makeIntForecastLease("secret/a", "critical", 1*time.Hour),
		makeIntForecastLease("secret/m", "warn", 24*time.Hour),
		makeIntForecastLease("secret/expired", "critical", -1*time.Hour),
	}

	opts := filter.ForecastOptions{Window: 72 * time.Hour, MinSeverity: "ok"}
	entries := filter.Forecast(leases, opts)

	if len(entries) != 3 {
		t.Fatalf("expected 3 forecast entries, got %d", len(entries))
	}
	if entries[0].Lease.Path != "secret/a" {
		t.Errorf("expected secret/a first, got %s", entries[0].Lease.Path)
	}
	if entries[2].Lease.Path != "secret/z" {
		t.Errorf("expected secret/z last, got %s", entries[2].Lease.Path)
	}
}

func TestForecast_FilterThenForecast_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntForecastLease("secret/prod/db", "critical", 2*time.Hour),
		makeIntForecastLease("secret/dev/db", "warn", 3*time.Hour),
		makeIntForecastLease("secret/prod/api", "ok", 10*time.Hour),
	}

	// First apply a path filter to prod only
	filtered := filter.Apply(leases, filter.Options{PathPrefix: "secret/prod"})

	opts := filter.ForecastOptions{Window: 72 * time.Hour, MinSeverity: "ok"}
	entries := filter.Forecast(filtered, opts)

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after filter+forecast, got %d", len(entries))
	}
	for _, e := range entries {
		if len(e.Lease.Path) < 12 || e.Lease.Path[:12] != "secret/prod/" {
			t.Errorf("unexpected path outside prod: %s", e.Lease.Path)
		}
	}
}
