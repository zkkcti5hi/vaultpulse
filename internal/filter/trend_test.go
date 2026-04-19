package filter

import (
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

func makeTrendLease(id, path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestTrend_CountsBySeverity(t *testing.T) {
	snap1 := []vault.SecretLease{
		makeTrendLease("a", "secret/a", "ok", 2*time.Hour),
		makeTrendLease("b", "secret/b", "warn", 30*time.Minute),
	}
	snap2 := []vault.SecretLease{
		makeTrendLease("a", "secret/a", "critical", 5*time.Minute),
		makeTrendLease("b", "secret/b", "critical", 3*time.Minute),
		makeTrendLease("c", "secret/c", "ok", 4*time.Hour),
	}
	points := Trend([][]vault.SecretLease{snap1, snap2})
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
	if points[0].Counts["ok"] != 1 || points[0].Counts["warn"] != 1 {
		t.Errorf("snap1 counts wrong: %v", points[0].Counts)
	}
	if points[1].Counts["critical"] != 2 || points[1].Counts["ok"] != 1 {
		t.Errorf("snap2 counts wrong: %v", points[1].Counts)
	}
}

func TestTrend_Empty(t *testing.T) {
	points := Trend(nil)
	if len(points) != 0 {
		t.Errorf("expected empty, got %d", len(points))
	}
}

func TestPrintTrend_ContainsHeaders(t *testing.T) {
	points := Trend([][]vault.SecretLease{{
		makeTrendLease("x", "secret/x", "warn", time.Hour),
	}})
	out := PrintTrend(points)
	for _, h := range []string{"TIME", "OK", "WARN", "CRITICAL"} {
		if !strings.Contains(out, h) {
			t.Errorf("missing header %q in output", h)
		}
	}
}

func TestPrintTrend_Empty(t *testing.T) {
	out := PrintTrend(nil)
	if !strings.Contains(out, "no trend data") {
		t.Errorf("expected no-data message, got: %s", out)
	}
}
