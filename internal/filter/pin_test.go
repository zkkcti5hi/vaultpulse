package filter_test

import (
	"testing"
	"time"

	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

func makePinLease(id, path, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
}

func TestPin_AddAndGet(t *testing.T) {
	store := filter.NewPinStore()
	l := makePinLease("id-1", "secret/a", "warning")
	store.Pin(l)
	got, ok := store.Get("id-1")
	if !ok {
		t.Fatal("expected lease to be found")
	}
	if got.Path != "secret/a" {
		t.Errorf("expected path secret/a, got %s", got.Path)
	}
}

func TestPin_GetNotFound(t *testing.T) {
	store := filter.NewPinStore()
	_, ok := store.Get("missing")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestPin_Unpin(t *testing.T) {
	store := filter.NewPinStore()
	l := makePinLease("id-2", "secret/b", "critical")
	store.Pin(l)
	if !store.Unpin("id-2") {
		t.Fatal("expected unpin to succeed")
	}
	if store.Len() != 0 {
		t.Errorf("expected 0 pins, got %d", store.Len())
	}
}

func TestPin_Unpin_NotFound(t *testing.T) {
	store := filter.NewPinStore()
	if store.Unpin("ghost") {
		t.Fatal("expected false for missing lease")
	}
}

func TestPin_List_Sorted(t *testing.T) {
	store := filter.NewPinStore()
	store.Pin(makePinLease("id-c", "secret/c", "ok"))
	store.Pin(makePinLease("id-a", "secret/a", "ok"))
	store.Pin(makePinLease("id-b", "secret/b", "ok"))
	list := store.List()
	if len(list) != 3 {
		t.Fatalf("expected 3, got %d", len(list))
	}
	if list[0].LeaseID != "id-a" || list[1].LeaseID != "id-b" || list[2].LeaseID != "id-c" {
		t.Errorf("unexpected order: %v", list)
	}
}
