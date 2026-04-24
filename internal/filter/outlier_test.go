package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

func makeOutlierLease(path string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  "lease/" + path,
		Path:     path,
		TTL:      ttl,
		Severity: "ok",
	}
}

func TestDetectOutliers_Empty(t *testing.T) {
	result := DetectOutliers(nil, 2.0)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestDetectOutliers_NoOutliers(t *testing.T) {
	leases := []vault.SecretLease{
		makeOutlierLease("a", 100*time.Second),
		makeOutlierLease("b", 110*time.Second),
		makeOutlierLease("c", 105*time.Second),
	}
	result := DetectOutliers(leases, 2.0)
	if len(result) != 0 {
		t.Errorf("expected no outliers, got %d", len(result))
	}
}

func TestDetectOutliers_DetectsLowTTL(t *testing.T) {
	leases := []vault.SecretLease{
		makeOutlierLease("a", 1000*time.Second),
		makeOutlierLease("b", 1000*time.Second),
		makeOutlierLease("c", 1000*time.Second),
		makeOutlierLease("d", 1000*time.Second),
		makeOutlierLease("outlier", 1*time.Second),
	}
	result := DetectOutliers(leases, 1.5)
	if len(result) == 0 {
		t.Fatal("expected at least one outlier")
	}
	if result[0].Lease.Path != "outlier" {
		t.Errorf("expected path 'outlier', got %s", result[0].Lease.Path)
	}
}

func TestDetectOutliers_DefaultMultiplier(t *testing.T) {
	leases := []vault.SecretLease{
		makeOutlierLease("a", 500*time.Second),
		makeOutlierLease("b", 500*time.Second),
		makeOutlierLease("c", 500*time.Second),
		makeOutlierLease("tiny", 1*time.Second),
	}
	// multiplier=0 should default to 2.0
	result := DetectOutliers(leases, 0)
	if len(result) == 0 {
		t.Fatal("expected outlier with default multiplier")
	}
}

func TestPrintOutliers_Empty(t *testing.T) {
	var buf bytes.Buffer
	PrintOutliers(nil, &buf)
	if buf.String() == "" {
		t.Error("expected output for empty set")
	}
}

func TestPrintOutliers_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeOutlierLease("a", 1000*time.Second),
		makeOutlierLease("b", 1000*time.Second),
		makeOutlierLease("c", 1000*time.Second),
		makeOutlierLease("low", 1*time.Second),
	}
	results := DetectOutliers(leases, 1.0)
	var buf bytes.Buffer
	PrintOutliers(results, &buf)
	out := buf.String()
	for _, hdr := range []string{"PATH", "TTL", "REASON"} {
		if !bytes.Contains([]byte(out), []byte(hdr)) {
			t.Errorf("missing header %q in output", hdr)
		}
	}
}
