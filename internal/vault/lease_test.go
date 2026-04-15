package vault

import (
	"testing"
	"time"
)

func makeLease(id string, ttl time.Duration) Lease {
	return Lease{
		ID:       id,
		Path:     "secret/" + id,
		TTL:      ttl,
		ExpireAt: time.Now().Add(ttl),
	}
}

func TestAnnotate_Severities(t *testing.T) {
	filter := ExpiryFilter{
		WarnBefore:     24 * time.Hour,
		CriticalBefore: 6 * time.Hour,
	}

	leases := []Lease{
		makeLease("ok-lease", 48*time.Hour),
		makeLease("warn-lease", 12*time.Hour),
		makeLease("critical-lease", 2*time.Hour),
	}

	annotated := Annotate(leases, filter)

	if len(annotated) != 3 {
		t.Fatalf("expected 3 annotated leases, got %d", len(annotated))
	}

	// After sorting, first should be critical.
	if annotated[0].Severity != SeverityCritical {
		t.Errorf("expected first lease to be critical, got %s", annotated[0].Severity)
	}
	if annotated[1].Severity != SeverityWarning {
		t.Errorf("expected second lease to be warning, got %s", annotated[1].Severity)
	}
	if annotated[2].Severity != SeverityOK {
		t.Errorf("expected third lease to be ok, got %s", annotated[2].Severity)
	}
}

func TestAnnotate_Empty(t *testing.T) {
	result := Annotate(nil, ExpiryFilter{
		WarnBefore:     time.Hour,
		CriticalBefore: 30 * time.Minute,
	})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
}

func TestAnnotate_AllCritical(t *testing.T) {
	filter := ExpiryFilter{
		WarnBefore:     24 * time.Hour,
		CriticalBefore: 6 * time.Hour,
	}

	leases := []Lease{
		makeLease("a", 5*time.Hour),
		makeLease("b", 3*time.Hour),
		makeLease("c", 1*time.Hour),
	}

	annotated := Annotate(leases, filter)
	for _, al := range annotated {
		if al.Severity != SeverityCritical {
			t.Errorf("lease %s: expected critical, got %s", al.ID, al.Severity)
		}
	}

	// Should be sorted by expiry ascending (soonest first within same severity).
	if annotated[0].ID != "c" {
		t.Errorf("expected soonest-expiring lease first, got %s", annotated[0].ID)
	}
}
