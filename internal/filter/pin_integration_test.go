package filter_test

import (
	"testing"
	"time"

	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

func makeIntPinLease(id, path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestPin_FilterThenPin_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntPinLease("a", "secret/db", "critical", 30*time.Minute),
		makeIntPinLease("b", "secret/api", "warning", 2*time.Hour),
		makeIntPinLease("c", "secret/cache", "ok", 24*time.Hour),
	}

	critical := filter.Apply(leases, filter.Options{MinSeverity: "critical"})

	store := filter.NewPinStore()
	for _, l := range critical {
		store.Pin(l)
	}

	if store.Len() != 1 {
		t.Fatalf("expected 1 pinned critical lease, got %d", store.Len())
	}
	got, ok := store.Get("a")
	if !ok {
		t.Fatal("expected lease 'a' to be pinned")
	}
	if got.Severity != "critical" {
		t.Errorf("expected critical severity, got %s", got.Severity)
	}
}

func TestPin_PinAndUnpin_Integration(t *testing.T) {
	store := filter.NewPinStore()
	for i, id := range []string{"x", "y", "z"} {
		store.Pin(makeIntPinLease(id, "secret/"+id, "warning", time.Duration(i+1)*time.Hour))
	}
	if store.Len() != 3 {
		t.Fatalf("expected 3, got %d", store.Len())
	}
	store.Unpin("y")
	list := store.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 after unpin, got %d", len(list))
	}
	if list[0].LeaseID != "x" || list[1].LeaseID != "z" {
		t.Errorf("unexpected list: %v", list)
	}
}
