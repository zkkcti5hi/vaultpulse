package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeCorrelationLease(path, severity string, tags []string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "lease-" + path,
		Path:      path,
		Severity:  severity,
		Tags:      tags,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
}

func TestCorrelate_ByPathPrefix(t *testing.T) {
	leases := []vault.SecretLease{
		makeCorrelationLease("secret/app/db", "critical", nil),
		makeCorrelationLease("secret/app/cache", "warn", nil),
		makeCorrelationLease("secret/other/svc", "ok", nil),
	}
	r := Correlate(leases, "path-prefix")
	if r.Field != "path-prefix" {
		t.Errorf("expected field path-prefix, got %s", r.Field)
	}
	if len(r.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(r.Groups))
	}
	if r.Groups[0].Key != "secret/app" {
		t.Errorf("expected key secret/app, got %s", r.Groups[0].Key)
	}
	if len(r.Groups[0].Leases) != 2 {
		t.Errorf("expected 2 leases in group, got %d", len(r.Groups[0].Leases))
	}
}

func TestCorrelate_BySeverity(t *testing.T) {
	leases := []vault.SecretLease{
		makeCorrelationLease("a/b", "critical", nil),
		makeCorrelationLease("c/d", "critical", nil),
		makeCorrelationLease("e/f", "ok", nil),
	}
	r := Correlate(leases, "severity")
	if len(r.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(r.Groups))
	}
	if r.Groups[0].Key != "critical" {
		t.Errorf("expected key critical, got %s", r.Groups[0].Key)
	}
}

func TestCorrelate_ByTag(t *testing.T) {
	leases := []vault.SecretLease{
		makeCorrelationLease("a/b", "warn", []string{"team:ops"}),
		makeCorrelationLease("c/d", "warn", []string{"team:ops"}),
		makeCorrelationLease("e/f", "ok", []string{"team:dev"}),
	}
	r := Correlate(leases, "tag")
	if len(r.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(r.Groups))
	}
}

func TestCorrelate_Empty(t *testing.T) {
	r := Correlate(nil, "severity")
	if len(r.Groups) != 0 {
		t.Errorf("expected no groups for empty input")
	}
}

func TestPrintCorrelation_ContainsKey(t *testing.T) {
	leases := []vault.SecretLease{
		makeCorrelationLease("secret/app/db", "critical", nil),
		makeCorrelationLease("secret/app/cache", "critical", nil),
	}
	r := Correlate(leases, "path-prefix")
	var buf bytes.Buffer
	PrintCorrelation(r, &buf)
	out := buf.String()
	if !strings.Contains(out, "secret/app") {
		t.Errorf("expected output to contain key 'secret/app', got:\n%s", out)
	}
	if !strings.Contains(out, "path-prefix") {
		t.Errorf("expected output to contain field name")
	}
}
