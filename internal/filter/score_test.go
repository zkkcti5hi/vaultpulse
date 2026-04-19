package filter

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeScoreLease(id, severity string, tags []string, labels map[string]string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      "secret/" + id,
		Severity:  severity,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Tags:      tags,
		Labels:    labels,
	}
}

func TestScore_OrderedByScore(t *testing.T) {
	leases := []vault.SecretLease{
		makeScoreLease("a", "info", nil, nil),
		makeScoreLease("b", "critical", nil, nil),
		makeScoreLease("c", "warning", nil, nil),
	}
	scored := Score(leases)
	if scored[0].Lease.LeaseID != "b" {
		t.Errorf("expected b first, got %s", scored[0].Lease.LeaseID)
	}
	if scored[1].Lease.LeaseID != "c" {
		t.Errorf("expected c second, got %s", scored[1].Lease.LeaseID)
	}
	if scored[2].Lease.LeaseID != "a" {
		t.Errorf("expected a third, got %s", scored[2].Lease.LeaseID)
	}
}

func TestScore_TagsAndLabelsAddPoints(t *testing.T) {
	base := makeScoreLease("x", "warning", nil, nil)
	rich := makeScoreLease("y", "warning", []string{"t1", "t2"}, map[string]string{"env": "prod"})
	scored := Score([]vault.SecretLease{base, rich})
	if scored[0].Lease.LeaseID != "y" {
		t.Errorf("expected y to rank higher due to tags/labels")
	}
	if scored[0].Score != 63 { // 50 + 10 + 3
		t.Errorf("unexpected score %d", scored[0].Score)
	}
}

func TestScore_Empty(t *testing.T) {
	scored := Score(nil)
	if len(scored) != 0 {
		t.Errorf("expected empty result")
	}
}

func TestScore_CriticalBaseScore(t *testing.T) {
	l := makeScoreLease("z", "critical", nil, nil)
	scored := Score([]vault.SecretLease{l})
	if scored[0].Score != 100 {
		t.Errorf("expected 100, got %d", scored[0].Score)
	}
}
