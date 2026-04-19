package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

func makeExpireLease(path string, expiresIn time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   path + "-id",
		Path:      path,
		ExpiresAt: time.Now().Add(expiresIn),
		Severity:  "warn",
	}
}

func TestFilterByExpireWindow_IncludesWithin(t *testing.T) {
	leases := []vault.SecretLease{
		makeExpireLease("secret/a", 10*time.Minute),
		makeExpireLease("secret/b", 2*time.Hour),
		makeExpireLease("secret/c", 25*time.Hour),
	}
	got := FilterByExpireWindow(leases, 1*time.Hour)
	if len(got) != 1 || got[0].Path != "secret/a" {
		t.Fatalf("expected secret/a only, got %v", got)
	}
}

func TestFilterByExpireWindow_ExcludesExpired(t *testing.T) {
	leases := []vault.SecretLease{
		{LeaseID: "x", Path: "secret/x", ExpiresAt: time.Now().Add(-1 * time.Minute), Severity: "critical"},
	}
	got := FilterByExpireWindow(leases, 1*time.Hour)
	if len(got) != 0 {
		t.Fatalf("expected no results, got %v", got)
	}
}

func TestGroupByExpireWindow_Keys(t *testing.T) {
	leases := []vault.SecretLease{
		makeExpireLease("secret/soon", 5*time.Minute),
		makeExpireLease("secret/later", 3*time.Hour),
	}
	windows := map[string]time.Duration{
		"1h":  1 * time.Hour,
		"24h": 24 * time.Hour,
	}
	groups := GroupByExpireWindow(leases, windows)
	if len(groups["1h"]) != 1 {
		t.Fatalf("expected 1 in 1h window, got %d", len(groups["1h"]))
	}
	if len(groups["24h"]) != 2 {
		t.Fatalf("expected 2 in 24h window, got %d", len(groups["24h"]))
	}
}

func TestPrintExpireWindows_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeExpireLease("secret/a", 10*time.Minute),
	}
	windows := map[string]time.Duration{"1h": time.Hour}
	groups := GroupByExpireWindow(leases, windows)
	var buf bytes.Buffer
	PrintExpireWindows(&buf, groups, []string{"1h"})
	out := buf.String()
	if !strings.Contains(out, "WINDOW") || !strings.Contains(out, "COUNT") {
		t.Fatalf("missing headers in output: %s", out)
	}
	if !strings.Contains(out, "secret/a") {
		t.Fatalf("missing path in output: %s", out)
	}
}
