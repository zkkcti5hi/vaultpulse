package filter_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeScheduleLease(id, path, severity string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func TestBuildSchedule_IncludesWithinWindow(t *testing.T) {
	leases := []vault.SecretLease{
		makeScheduleLease("a", "secret/a", "critical", 30*time.Minute),
		makeScheduleLease("b", "secret/b", "warn", 2*time.Hour),
	}
	entries := filter.BuildSchedule(leases, 1*time.Hour)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].LeaseID != "a" {
		t.Errorf("expected lease 'a', got %s", entries[0].LeaseID)
	}
}

func TestBuildSchedule_ExcludesExpired(t *testing.T) {
	leases := []vault.SecretLease{
		makeScheduleLease("x", "secret/x", "critical", -5*time.Minute),
	}
	entries := filter.BuildSchedule(leases, 1*time.Hour)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestBuildSchedule_Empty(t *testing.T) {
	entries := filter.BuildSchedule(nil, 1*time.Hour)
	if len(entries) != 0 {
		t.Errorf("expected empty schedule")
	}
}

func TestFilterScheduleByMinSeverity(t *testing.T) {
	entries := []filter.ScheduleEntry{
		{LeaseID: "1", Severity: "critical"},
		{LeaseID: "2", Severity: "warn"},
		{LeaseID: "3", Severity: "ok"},
	}
	out := filter.FilterScheduleByMinSeverity(entries, "warn")
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestScheduleEntry_String(t *testing.T) {
	e := filter.ScheduleEntry{
		LeaseID:  "abc",
		Path:     "secret/test",
		Severity: "critical",
		NotifyAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	s := e.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
