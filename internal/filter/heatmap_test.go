package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeHeatmapLease(id, path, severity string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func TestHeatmap_CountsByWindow(t *testing.T) {
	leases := []vault.SecretLease{
		makeHeatmapLease("l1", "secret/a", "critical", 30*time.Minute),
		makeHeatmapLease("l2", "secret/b", "warn", 3*time.Hour),
		makeHeatmapLease("l3", "secret/c", "critical", 30*time.Minute),
	}
	opts := DefaultHeatmapOptions()
	cells := Heatmap(leases, opts)

	critIn1h := 0
	for _, c := range cells {
		if c.Window == "1h" && c.Severity == "critical" {
			critIn1h = c.Count
		}
	}
	if critIn1h != 2 {
		t.Errorf("expected 2 critical leases in 1h window, got %d", critIn1h)
	}
}

func TestHeatmap_ExcludesExpired(t *testing.T) {
	leases := []vault.SecretLease{
		makeHeatmapLease("expired", "secret/x", "critical", -1*time.Hour),
	}
	cells := Heatmap(leases, DefaultHeatmapOptions())
	for _, c := range cells {
		if c.Severity == "critical" {
			t.Errorf("expired lease should not appear in heatmap")
		}
	}
}

func TestHeatmap_Empty(t *testing.T) {
	cells := Heatmap(nil, DefaultHeatmapOptions())
	if len(cells) != 0 {
		t.Errorf("expected empty cells for nil input")
	}
}

func TestHeatmap_WarnIn6hNotIn1h(t *testing.T) {
	leases := []vault.SecretLease{
		makeHeatmapLease("w1", "secret/w", "warn", 3*time.Hour),
	}
	cells := Heatmap(leases, DefaultHeatmapOptions())

	for _, c := range cells {
		if c.Window == "1h" && c.Severity == "warn" {
			t.Errorf("warn lease expiring in 3h should not appear in 1h window")
		}
	}
	found6h := false
	for _, c := range cells {
		if c.Window == "6h" && c.Severity == "warn" && c.Count == 1 {
			found6h = true
		}
	}
	if !found6h {
		t.Errorf("expected warn lease in 6h window")
	}
}

func TestPrintHeatmap_ContainsHeaders(t *testing.T) {
	cells := []HeatmapCell{
		{Window: "1h", Severity: "critical", Count: 2},
	}
	var buf bytes.Buffer
	PrintHeatmap(cells, &buf)
	out := buf.String()
	for _, hdr := range []string{"WINDOW", "SEVERITY", "COUNT"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
	if !strings.Contains(out, "1h") || !strings.Contains(out, "critical") {
		t.Errorf("expected cell data in output")
	}
}
