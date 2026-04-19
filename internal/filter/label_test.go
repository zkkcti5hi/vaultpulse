package filter

import (
	"testing"
	"time"

	"github.com/nicholasgasior/vaultpulse/internal/vault"
)

func makeLabelLease(id, path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestLabel_AddAndGet(t *testing.T) {
	s := NewLabelStore()
	s.Add("lease-1", "production")
	s.Add("lease-1", "team-a")
	labels := s.Get("lease-1")
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}
}

func TestLabel_NoDuplicates(t *testing.T) {
	s := NewLabelStore()
	s.Add("lease-1", "production")
	s.Add("lease-1", "production")
	if len(s.Get("lease-1")) != 1 {
		t.Fatal("expected duplicate label to be ignored")
	}
}

func TestLabel_Remove(t *testing.T) {
	s := NewLabelStore()
	s.Add("lease-1", "production")
	s.Add("lease-1", "staging")
	s.Remove("lease-1", "staging")
	labels := s.Get("lease-1")
	if len(labels) != 1 || labels[0] != "production" {
		t.Fatalf("unexpected labels after remove: %v", labels)
	}
}

func TestLabel_FilterByLabel(t *testing.T) {
	s := NewLabelStore()
	leases := []vault.SecretLease{
		makeLabelLease("lease-1", "secret/a"),
		makeLabelLease("lease-2", "secret/b"),
		makeLabelLease("lease-3", "secret/c"),
	}
	s.Add("lease-1", "critical")
	s.Add("lease-3", "critical")
	result := s.FilterByLabel(leases, "critical")
	if len(result) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(result))
	}
}

func TestLabel_FilterByLabel_NoMatch(t *testing.T) {
	s := NewLabelStore()
	leases := []vault.SecretLease{makeLabelLease("lease-1", "secret/a")}
	result := s.FilterByLabel(leases, "ghost")
	if len(result) != 0 {
		t.Fatal("expected no results")
	}
}
