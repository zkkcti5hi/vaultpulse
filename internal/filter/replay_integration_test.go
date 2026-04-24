package filter_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeIntReplayLease(id, severity string, ttl int) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     "secret/" + id,
		Severity: severity,
		TTL:      ttl,
	}
}

func TestReplay_RecordThenQuery_Integration(t *testing.T) {
	store := filter.NewReplayStore()
	base := time.Now()

	leases1 := []vault.SecretLease{makeIntReplayLease("a", "ok", 3600)}
	leases2 := []vault.SecretLease{
		makeIntReplayLease("b", "warn", 1800),
		makeIntReplayLease("c", "critical", 300),
	}

	store.Record(base.Add(-20*time.Minute), leases1)
	store.Record(base.Add(-5*time.Minute), leases2)

	// Query near the second event
	e, ok := store.At(base.Add(-4 * time.Minute))
	if !ok {
		t.Fatal("expected event")
	}
	if len(e.Leases) != 2 {
		t.Errorf("expected 2 leases, got %d", len(e.Leases))
	}
}

func TestReplay_AllOrdered_Integration(t *testing.T) {
	store := filter.NewReplayStore()
	base := time.Now()

	for i := 9; i >= 0; i-- {
		store.Record(base.Add(time.Duration(i)*time.Minute), []vault.SecretLease{
			makeIntReplayLease(fmt.Sprintf("l%d", i), "ok", 3600),
		})
	}

	all := store.All()
	if len(all) != 10 {
		t.Fatalf("expected 10 events, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i].At.Before(all[i-1].At) {
			t.Errorf("not sorted at index %d", i)
		}
	}
}
