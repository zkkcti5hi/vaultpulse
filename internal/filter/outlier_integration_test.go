package filter_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeIntOutlierLease(path string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  "lease/" + path,
		Path:     path,
		TTL:      ttl,
		Severity: "ok",
	}
}

func TestOutlier_FilterThenDetect_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntOutlierLease("secret/prod/db", 3600*time.Second),
		makeIntOutlierLease("secret/prod/api", 3500*time.Second),
		makeIntOutlierLease("secret/prod/cache", 3700*time.Second),
		makeIntOutlierLease("secret/prod/mq", 3600*time.Second),
		makeIntOutlierLease("secret/prod/short", 10*time.Second),
	}

	// First filter to prod prefix
	filtered := filter.Apply(leases, filter.Options{PathPrefix: "secret/prod"})
	if len(filtered) != 5 {
		t.Fatalf("expected 5 filtered leases, got %d", len(filtered))
	}

	// Then detect outliers
	outliers := filter.DetectOutliers(filtered, 1.5)
	if len(outliers) == 0 {
		t.Fatal("expected at least one outlier after filter")
	}
	if outliers[0].Lease.Path != "secret/prod/short" {
		t.Errorf("expected 'secret/prod/short' as outlier, got %s", outliers[0].Lease.Path)
	}
}

func TestOutlier_EmptyLeases_Integration(t *testing.T) {
	results := filter.DetectOutliers([]vault.SecretLease{}, 2.0)
	if results != nil {
		t.Errorf("expected nil for empty input, got %v", results)
	}
}

func TestOutlier_AllSameTTL_NoOutliers_Integration(t *testing.T) {
	leases := make([]vault.SecretLease, 10)
	for i := range leases {
		leases[i] = makeIntOutlierLease("secret/uniform", 600*time.Second)
	}
	results := filter.DetectOutliers(leases, 2.0)
	if len(results) != 0 {
		t.Errorf("expected no outliers for uniform TTL, got %d", len(results))
	}
}
