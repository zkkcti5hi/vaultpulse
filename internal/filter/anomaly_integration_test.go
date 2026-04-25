package filter_test

import (
	"testing"
	"time"

	"github.com/user/vaultpulse/internal/filter"
	"github.com/user/vaultpulse/internal/vault"
)

func makeIntAnomalyLease(path, severity string, expiresIn time.Duration, seenAgo time.Duration) vault.SecretLease {
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

func TestAnomaly_FilterThenDetect_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntAnomalyLease("secret/a", "critical", 2*time.Minute, -1),
		makeIntAnomalyLease("secret/b", "warn", 1*time.Hour, -1),
		makeIntAnomalyLease("secret/c", "ok", 6*time.Hour, -1),
	}

	// First apply a severity filter to keep only critical/warn.
	filtered := filter.Apply(leases, filter.Options{MinSeverity: "warn"})

	opts := filter.DefaultAnomalyOptions()
	results := filter.DetectAnomalies(filtered, opts)

	if len(results) != 1 {
		t.Fatalf("expected 1 anomaly after filter, got %d", len(results))
	}
	if results[0].Lease.Path != "secret/a" {
		t.Errorf("expected secret/a, got %s", results[0].Lease.Path)
	}
}

func TestAnomaly_EmptyLeases_Integration(t *testing.T) {
	results := filter.DetectAnomalies([]vault.SecretLease{}, filter.DefaultAnomalyOptions())
	if len(results) != 0 {
		t.Errorf("expected no anomalies for empty input")
	}
}
