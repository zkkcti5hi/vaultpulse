package filter_test

import (
	"testing"
	"time"

	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

func makeLease(path, severity string) vault.SecretLease {
	return vault.SecretLease{
		Path:      path,
		LeaseID:   path + "/lease",
		Severity:  severity,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestApply_NoFilter(t *testing.T) {
	leases := []vault.SecretLease{
		makeLease("secret/a", "ok"),
		makeLease("secret/b", "warning"),
	}
	got := filter.Apply(leases, filter.Options{})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestApply_SeverityFilter(t *testing.T) {
	leases := []vault.SecretLease{
		makeLease("secret/a", "ok"),
		makeLease("secret/b", "warning"),
		makeLease("secret/c", "critical"),
	}
	got := filter.Apply(leases, filter.Options{Severity: "warning"})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	for _, l := range got {
		if l.Severity == "ok" {
			t.Errorf("unexpected ok lease in results")
		}
	}
}

func TestApply_PathPrefixFilter(t *testing.T) {
	leases := []vault.SecretLease{
		makeLease("secret/app1/db", "ok"),
		makeLease("secret/app2/db", "warning"),
		makeLease("secret/app1/api", "critical"),
	}
	got := filter.Apply(leases, filter.Options{PathPrefix: "secret/app1"})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestApply_CombinedFilter(t *testing.T) {
	leases := []vault.SecretLease{
		makeLease("secret/app1/db", "ok"),
		makeLease("secret/app1/api", "critical"),
		makeLease("secret/app2/db", "critical"),
	}
	got := filter.Apply(leases, filter.Options{PathPrefix: "secret/app1", Severity: "critical"})
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].Path != "secret/app1/api" {
		t.Errorf("unexpected path %s", got[0].Path)
	}
}

func TestApply_EmptyLeases(t *testing.T) {
	got := filter.Apply(nil, filter.Options{Severity: "critical"})
	if len(got) != 0 {
		t.Fatalf("expected 0, got %d", len(got))
	}
}
