package filter_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeIntCorrelationLease(path, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "lease-" + path,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
}

func TestCorrelation_FilterThenCorrelate_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntCorrelationLease("secret/app/db", "critical"),
		makeIntCorrelationLease("secret/app/cache", "critical"),
		makeIntCorrelationLease("secret/other/svc", "ok"),
		makeIntCorrelationLease("secret/other/job", "ok"),
	}

	// Filter to critical only, then correlate.
	filtered := filter.Apply(leases, filter.Options{MinSeverity: "critical"})
	r := filter.Correlate(filtered, "path-prefix")

	if len(r.Groups) != 1 {
		t.Fatalf("expected 1 group after filter+correlate, got %d", len(r.Groups))
	}
	if r.Groups[0].Key != "secret/app" {
		t.Errorf("expected group key secret/app, got %s", r.Groups[0].Key)
	}
}

func TestCorrelation_EmptyLeases_Integration(t *testing.T) {
	r := filter.Correlate(nil, "severity")
	if len(r.Groups) != 0 {
		t.Errorf("expected no groups for nil input")
	}
}

func TestCorrelation_SortedByGroupSize_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntCorrelationLease("secret/big/a", "warn"),
		makeIntCorrelationLease("secret/big/b", "warn"),
		makeIntCorrelationLease("secret/big/c", "warn"),
		makeIntCorrelationLease("secret/small/x", "warn"),
		makeIntCorrelationLease("secret/small/y", "warn"),
	}
	r := filter.Correlate(leases, "path-prefix")
	if len(r.Groups) < 2 {
		t.Fatalf("expected at least 2 groups, got %d", len(r.Groups))
	}
	if len(r.Groups[0].Leases) < len(r.Groups[1].Leases) {
		t.Errorf("expected groups sorted by size descending")
	}
}
