package filter

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

func makeLeaseSort(path, severity string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func TestSort_ByExpiry_Ascending(t *testing.T) {
	leases := []vault.SecretLease{
		makeLeaseSort("c", "ok", 3*time.Hour),
		makeLeaseSort("a", "ok", 1*time.Hour),
		makeLeaseSort("b", "ok", 2*time.Hour),
	}
	out := Sort(leases, SortOptions{Field: SortByExpiry, Order: Ascending})
	if out[0].Path != "a" || out[1].Path != "b" || out[2].Path != "c" {
		t.Fatalf("expected a,b,c got %s,%s,%s", out[0].Path, out[1].Path, out[2].Path)
	}
}

func TestSort_ByExpiry_Descending(t *testing.T) {
	leases := []vault.SecretLease{
		makeLeaseSort("a", "ok", 1*time.Hour),
		makeLeaseSort("c", "ok", 3*time.Hour),
	}
	out := Sort(leases, SortOptions{Field: SortByExpiry, Order: Descending})
	if out[0].Path != "a" {
		t.Fatalf("expected descending to put soonest-expiring last, got first=%s", out[0].Path)
	}
}

func TestSort_BySeverity(t *testing.T) {
	leases := []vault.SecretLease{
		makeLeaseSort("x", "warning", time.Hour),
		makeLeaseSort("y", "critical", time.Hour),
		makeLeaseSort("z", "ok", time.Hour),
	}
	out := Sort(leases, SortOptions{Field: SortBySeverity, Order: Ascending})
	if out[0].Severity != "critical" {
		t.Fatalf("expected critical first, got %s", out[0].Severity)
	}
}

func TestSort_ByPath(t *testing.T) {
	leases := []vault.SecretLease{
		makeLeaseSort("secret/z", "ok", time.Hour),
		makeLeaseSort("secret/a", "ok", time.Hour),
	}
	out := Sort(leases, SortOptions{Field: SortByPath, Order: Ascending})
	if out[0].Path != "secret/a" {
		t.Fatalf("expected secret/a first, got %s", out[0].Path)
	}
}

func TestSort_EmptyField_PreservesOrder(t *testing.T) {
	leases := []vault.SecretLease{
		makeLeaseSort("b", "ok", time.Hour),
		makeLeaseSort("a", "ok", time.Hour),
	}
	out := Sort(leases, SortOptions{})
	if out[0].Path != "b" {
		t.Fatal("expected original order preserved")
	}
}
