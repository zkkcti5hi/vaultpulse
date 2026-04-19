package filter_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeIntExpireLease(path string, expiresIn time.Duration, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   path + "-id",
		Path:      path,
		ExpiresAt: time.Now().Add(expiresIn),
		Severity:  severity,
	}
}

func TestExpire_FilterThenGroup_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntExpireLease("secret/a", 30*time.Minute, "critical"),
		makeIntExpireLease("secret/b", 3*time.Hour, "warn"),
		makeIntExpireLease("secret/c", 10*time.Hour, "ok"),
	}

	// First filter to critical only
	critical := filter.Apply(leases, filter.Options{MinSeverity: "critical"})
	if len(critical) != 1 {
		t.Fatalf("expected 1 critical lease, got %d", len(critical))
	}

	// Then group by window — critical lease should appear in 1h window
	windows := map[string]time.Duration{"1h": time.Hour, "6h": 6 * time.Hour}
	groups := filter.GroupByExpireWindow(critical, windows)
	if len(groups["1h"]) != 1 {
		t.Fatalf("expected critical lease in 1h window, got %d", len(groups["1h"]))
	}
	if len(groups["6h"]) != 1 {
		t.Fatalf("expected critical lease also in 6h window, got %d", len(groups["6h"]))
	}
}

func TestExpire_EmptyLeases_Integration(t *testing.T) {
	groups := filter.GroupByExpireWindow(nil, map[string]time.Duration{"1h": time.Hour})
	if len(groups["1h"]) != 0 {
		t.Fatalf("expected empty window group, got %d", len(groups["1h"]))
	}
}
