package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/coolguy/vaultpulse/internal/vault"
)

func makeDecayLease(id, path, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		Severity: severity,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
}

func TestDecay_NoSeenAt_ZeroAge(t *testing.T) {
	leases := []vault.SecretLease{
		makeDecayLease("id-1", "secret/a", "critical"),
	}

	entries := Decay(leases, map[string]time.Time{})

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Age != 0 {
		t.Errorf("expected zero age, got %v", entries[0].Age)
	}
	if entries[0].DecayScore != 100.0 {
		t.Errorf("expected decay score 100.0, got %.2f", entries[0].DecayScore)
	}
}

func TestDecay_AgeIncreasesScore(t *testing.T) {
	seenAt := map[string]time.Time{
		"id-2": time.Now().Add(-2 * time.Hour),
	}
	leases := []vault.SecretLease{
		makeDecayLease("id-2", "secret/b", "warn"),
	}

	entries := Decay(leases, seenAt)

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// base 50 + 2h * 2.0 = 54.0
	if entries[0].DecayScore < 53.9 || entries[0].DecayScore > 54.1 {
		t.Errorf("expected decay score ~54.0, got %.2f", entries[0].DecayScore)
	}
}

func TestDecay_SortedByScoreDescending(t *testing.T) {
	seenAt := map[string]time.Time{
		"id-ok":   time.Now().Add(-1 * time.Hour),
		"id-crit": time.Now().Add(-1 * time.Hour),
	}
	leases := []vault.SecretLease{
		makeDecayLease("id-ok", "secret/ok", "ok"),
		makeDecayLease("id-crit", "secret/crit", "critical"),
	}

	entries := Decay(leases, seenAt)

	if entries[0].Severity != "critical" {
		t.Errorf("expected critical first, got %s", entries[0].Severity)
	}
}

func TestDecay_Empty(t *testing.T) {
	entries := Decay([]vault.SecretLease{}, map[string]time.Time{})
	if len(entries) != 0 {
		t.Errorf("expected empty result, got %d entries", len(entries))
	}
}

func TestPrintDecay_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeDecayLease("id-x", "secret/x", "warn"),
	}
	entries := Decay(leases, map[string]time.Time{})

	var buf bytes.Buffer
	PrintDecay(entries, &buf)
	out := buf.String()

	for _, hdr := range []string{"LEASE ID", "PATH", "SEVERITY", "AGE", "DECAY SCORE"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}
