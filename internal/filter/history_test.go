package filter_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeHistoryLease(id string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      "secret/" + id,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestHistory_RecordAndLen(t *testing.T) {
	h := filter.NewHistory(5)
	if h.Len() != 0 {
		t.Fatalf("expected 0, got %d", h.Len())
	}
	h.Record([]vault.SecretLease{makeHistoryLease("a")})
	h.Record([]vault.SecretLease{makeHistoryLease("b")})
	if h.Len() != 2 {
		t.Fatalf("expected 2, got %d", h.Len())
	}
}

func TestHistory_Latest(t *testing.T) {
	h := filter.NewHistory(5)
	_, ok := h.Latest()
	if ok {
		t.Fatal("expected false on empty history")
	}
	h.Record([]vault.SecretLease{makeHistoryLease("x")})
	h.Record([]vault.SecretLease{makeHistoryLease("y"), makeHistoryLease("z")})
	entry, ok := h.Latest()
	if !ok {
		t.Fatal("expected entry")
	}
	if len(entry.Leases) != 2 {
		t.Fatalf("expected 2 leases in latest, got %d", len(entry.Leases))
	}
}

func TestHistory_RollingWindow(t *testing.T) {
	h := filter.NewHistory(3)
	for i := 0; i < 5; i++ {
		h.Record([]vault.SecretLease{makeHistoryLease("id")})
	}
	if h.Len() != 3 {
		t.Fatalf("expected max 3, got %d", h.Len())
	}
}

func TestHistory_All_Order(t *testing.T) {
	h := filter.NewHistory(5)
	h.Record([]vault.SecretLease{makeHistoryLease("first")})
	h.Record([]vault.SecretLease{makeHistoryLease("second")})
	all := h.All()
	if all[0].Leases[0].LeaseID != "first" {
		t.Errorf("expected first snapshot to be oldest")
	}
	if all[1].Leases[0].LeaseID != "second" {
		t.Errorf("expected second snapshot to be newest")
	}
}

func TestHistory_DefaultMax(t *testing.T) {
	h := filter.NewHistory(0)
	for i := 0; i < 12; i++ {
		h.Record([]vault.SecretLease{})
	}
	if h.Len() != 10 {
		t.Fatalf("expected default max 10, got %d", h.Len())
	}
}
