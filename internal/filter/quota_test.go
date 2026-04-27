package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeQuotaLease(path, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "lease/" + path,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
}

func TestApplyQuota_NoViolations(t *testing.T) {
	leases := []vault.SecretLease{
		makeQuotaLease("secret/app/db", "ok"),
		makeQuotaLease("secret/app/api", "warn"),
	}
	opts := DefaultQuotaOptions()
	violations := ApplyQuota(leases, opts)
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}
}

func TestApplyQuota_PathViolation(t *testing.T) {
	var leases []vault.SecretLease
	for i := 0; i < 5; i++ {
		leases = append(leases, makeQuotaLease("secret/app/key", "ok"))
	}
	opts := QuotaOptions{MaxPerPath: 3, MaxPerSeverity: 100}
	violations := ApplyQuota(leases, opts)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Dimension != "path" {
		t.Errorf("expected dimension=path, got %q", violations[0].Dimension)
	}
	if violations[0].Count != 5 {
		t.Errorf("expected count=5, got %d", violations[0].Count)
	}
}

func TestApplyQuota_SeverityViolation(t *testing.T) {
	var leases []vault.SecretLease
	for i := 0; i < 4; i++ {
		leases = append(leases, makeQuotaLease("secret/svc"+string(rune('a'+i)), "critical"))
	}
	opts := QuotaOptions{MaxPerPath: 100, MaxPerSeverity: 2}
	violations := ApplyQuota(leases, opts)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Dimension != "severity" {
		t.Errorf("expected dimension=severity, got %q", violations[0].Dimension)
	}
}

func TestApplyQuota_Empty(t *testing.T) {
	violations := ApplyQuota(nil, DefaultQuotaOptions())
	if len(violations) != 0 {
		t.Fatalf("expected no violations on empty input")
	}
}

func TestPrintQuota_ContainsHeaders(t *testing.T) {
	violations := []QuotaViolation{
		{Dimension: "path", Key: "secret/app", Count: 25, Limit: 20},
	}
	var buf bytes.Buffer
	PrintQuota(violations, &buf)
	out := buf.String()
	for _, hdr := range []string{"DIMENSION", "KEY", "COUNT", "LIMIT"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestPrintQuota_NoViolations(t *testing.T) {
	var buf bytes.Buffer
	PrintQuota(nil, &buf)
	if !strings.Contains(buf.String(), "No quota violations") {
		t.Error("expected no-violations message")
	}
}
