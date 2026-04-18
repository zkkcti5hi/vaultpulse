package filter_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeCompareLease(id, path, severity string) vault.Lease {
	return vault.Lease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestCompare_Added(t *testing.T) {
	before := []vault.Lease{makeCompareLease("a", "secret/a", "ok")}
	after := []vault.Lease{
		makeCompareLease("a", "secret/a", "ok"),
		makeCompareLease("b", "secret/b", "warning"),
	}
	r := filter.Compare(before, after)
	if len(r.Added) != 1 || r.Added[0].LeaseID != "b" {
		t.Errorf("expected 1 added lease 'b', got %+v", r.Added)
	}
	if len(r.Removed) != 0 || len(r.Changed) != 0 {
		t.Errorf("unexpected removed/changed: %s", r)
	}
}

func TestCompare_Removed(t *testing.T) {
	before := []vault.Lease{
		makeCompareLease("a", "secret/a", "ok"),
		makeCompareLease("b", "secret/b", "warning"),
	}
	after := []vault.Lease{makeCompareLease("a", "secret/a", "ok")}
	r := filter.Compare(before, after)
	if len(r.Removed) != 1 || r.Removed[0].LeaseID != "b" {
		t.Errorf("expected 1 removed lease 'b', got %+v", r.Removed)
	}
}

func TestCompare_Changed(t *testing.T) {
	before := []vault.Lease{makeCompareLease("a", "secret/a", "ok")}
	after := []vault.Lease{makeCompareLease("a", "secret/a", "critical")}
	r := filter.Compare(before, after)
	if len(r.Changed) != 1 || r.Changed[0].LeaseID != "a" {
		t.Errorf("expected 1 changed lease 'a', got %+v", r.Changed)
	}
}

func TestCompare_NoChange(t *testing.T) {
	leases := []vault.Lease{makeCompareLease("a", "secret/a", "ok")}
	r := filter.Compare(leases, leases)
	if len(r.Added)+len(r.Removed)+len(r.Changed) != 0 {
		t.Errorf("expected no diff, got %s", r)
	}
}

func TestCompare_Empty(t *testing.T) {
	r := filter.Compare(nil, nil)
	if r.String() != "Added: 0, Removed: 0, Changed: 0" {
		t.Errorf("unexpected string: %s", r)
	}
}
