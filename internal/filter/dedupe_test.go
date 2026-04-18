package filter_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeDedupeLease(id, path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestDedupe_NoDuplicates(t *testing.T) {
	leases := []vault.SecretLease{
		makeDedupeLease("id-1", "secret/a"),
		makeDedupeLease("id-2", "secret/b"),
	}
	result := filter.Dedupe(leases)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestDedupe_RemovesDuplicates(t *testing.T) {
	leases := []vault.SecretLease{
		makeDedupeLease("id-1", "secret/a"),
		makeDedupeLease("id-1", "secret/a-copy"),
		makeDedupeLease("id-2", "secret/b"),
	}
	result := filter.Dedupe(leases)
	if len(result) != 2 {
		t.Fatalf("expected 2 after dedupe, got %d", len(result))
	}
	if result[0].LeaseID != "id-1" || result[1].LeaseID != "id-2" {
		t.Errorf("unexpected lease IDs: %v", result)
	}
}

func TestDedupe_Empty(t *testing.T) {
	result := filter.Dedupe(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty, got %d", len(result))
	}
}

func TestDedupe_KeepsFirstOccurrence(t *testing.T) {
	leases := []vault.SecretLease{
		makeDedupeLease("id-1", "secret/first"),
		makeDedupeLease("id-1", "secret/second"),
	}
	result := filter.Dedupe(leases)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].Path != "secret/first" {
		t.Errorf("expected first occurrence to be kept, got path %s", result[0].Path)
	}
}
