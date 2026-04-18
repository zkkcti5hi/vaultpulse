package filter

import (
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

func makeSumLease(path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		Path:      path,
		Severity:  severity,
		TTL:       ttl,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestSummarize_Counts(t *testing.T) {
	leases := []vault.SecretLease{
		makeSumLease("secret/a", "critical", -1*time.Second),
		makeSumLease("secret/b", "critical", 30*time.Second),
		makeSumLease("secret/c", "warning", 5*time.Minute),
		makeSumLease("secret/d", "ok", time.Hour),
	}
	s := Summarize(leases)
	if s.Total != 4 {
		t.Errorf("expected total 4, got %d", s.Total)
	}
	if s.BySeverity["critical"] != 2 {
		t.Errorf("expected 2 critical, got %d", s.BySeverity["critical"])
	}
	if s.BySeverity["warning"] != 1 {
		t.Errorf("expected 1 warning, got %d", s.BySeverity["warning"])
	}
	if s.ExpiredCount != 1 {
		t.Errorf("expected 1 expired, got %d", s.ExpiredCount)
	}
}

func TestSummarize_CriticalPaths(t *testing.T) {
	leases := []vault.SecretLease{
		makeSumLease("secret/db/pass", "critical", time.Second),
		makeSumLease("secret/db/user", "critical", time.Second),
		makeSumLease("secret/app/key", "critical", time.Second),
	}
	s := Summarize(leases)
	if len(s.CriticalPaths) != 2 {
		t.Errorf("expected 2 unique critical prefixes, got %d: %v", len(s.CriticalPaths), s.CriticalPaths)
	}
}

func TestSummarize_Empty(t *testing.T) {
	s := Summarize(nil)
	if s.Total != 0 || s.ExpiredCount != 0 {
		t.Error("expected zero summary for nil input")
	}
}

func TestSummary_String(t *testing.T) {
	leases := []vault.SecretLease{
		makeSumLease("secret/a", "critical", time.Second),
		makeSumLease("secret/b", "ok", time.Hour),
	}
	s := Summarize(leases)
	out := s.String()
	if !strings.Contains(out, "total=2") {
		t.Errorf("unexpected summary string: %s", out)
	}
}
