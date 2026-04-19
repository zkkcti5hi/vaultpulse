package filter

import (
	"testing"
	"time"

	"github.com/vaultpulse/internal/vault"
)

func makeSuppressLease(id, path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestSuppress_IndefinitelyHides(t *testing.T) {
	s := NewSuppressStore()
	s.Suppress("lease-1", time.Time{})
	if !s.IsSuppressed("lease-1") {
		t.Fatal("expected lease-1 to be suppressed")
	}
}

func TestSuppress_ExpiredAllowsThrough(t *testing.T) {
	s := NewSuppressStore()
	s.Suppress("lease-2", time.Now().Add(-time.Second))
	if s.IsSuppressed("lease-2") {
		t.Fatal("expected lease-2 to be expired and not suppressed")
	}
}

func TestSuppress_FutureUntilSuppresses(t *testing.T) {
	s := NewSuppressStore()
	s.Suppress("lease-3", time.Now().Add(time.Hour))
	if !s.IsSuppressed("lease-3") {
		t.Fatal("expected lease-3 to be suppressed")
	}
}

func TestUnsuppress_RemovesEntry(t *testing.T) {
	s := NewSuppressStore()
	s.Suppress("lease-4", time.Time{})
	ok := s.Unsuppress("lease-4")
	if !ok {
		t.Fatal("expected Unsuppress to return true")
	}
	if s.IsSuppressed("lease-4") {
		t.Fatal("expected lease-4 to no longer be suppressed")
	}
}

func TestUnsuppress_NotFound(t *testing.T) {
	s := NewSuppressStore()
	if s.Unsuppress("missing") {
		t.Fatal("expected false for missing lease")
	}
}

func TestApplySuppress_FiltersLeases(t *testing.T) {
	s := NewSuppressStore()
	s.Suppress("id-1", time.Time{})
	leases := []vault.SecretLease{
		makeSuppressLease("id-1", "secret/a"),
		makeSuppressLease("id-2", "secret/b"),
	}
	out := s.ApplySuppress(leases)
	if len(out) != 1 || out[0].LeaseID != "id-2" {
		t.Fatalf("expected only id-2, got %v", out)
	}
}

func TestSuppress_List(t *testing.T) {
	s := NewSuppressStore()
	s.Suppress("a", time.Time{})
	s.Suppress("b", time.Now().Add(time.Hour))
	list := s.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
}
