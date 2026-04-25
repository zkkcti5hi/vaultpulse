package filter_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeIntHeatmapLease(id, path, severity string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func TestHeatmap_FilterThenHeatmap_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntHeatmapLease("a1", "prod/db", "critical", 20*time.Minute),
		makeIntHeatmapLease("a2", "prod/api", "warn", 4*time.Hour),
		makeIntHeatmapLease("a3", "staging/db", "ok", 50*time.Hour),
		makeIntHeatmapLease("a4", "prod/cache", "critical", 45*time.Minute),
	}

	// Filter to prod/ only
	filtered := filter.Apply(leases, filter.Options{PathPrefix: "prod/"})

	opts := filter.DefaultHeatmapOptions()
	cells := filter.Heatmap(filtered, opts)

	critIn1h := 0
	for _, c := range cells {
		if c.Window == "1h" && c.Severity == "critical" {
			critIn1h = c.Count
		}
	}
	if critIn1h != 2 {
		t.Errorf("expected 2 critical leases in 1h window after filter, got %d", critIn1h)
	}
}

func TestHeatmap_EmptyLeases_Integration(t *testing.T) {
	cells := filter.Heatmap(nil, filter.DefaultHeatmapOptions())
	if len(cells) != 0 {
		t.Errorf("expected no cells for empty input")
	}
}
