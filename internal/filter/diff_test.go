package filter_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeDiffLease(id, path, severity string, ttl int) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		Severity: severity,
		TTL:      ttl,
		ExpireAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}
}

func TestDiff_Added(t *testing.T) {
	before := []vault.SecretLease{makeDiffLease("a", "sec/a", "ok", 3600)}
	after := []vault.SecretLease{
		makeDiffLease("a", "sec/a", "ok", 3600),
		makeDiffLease("b", "sec/b", "warn", 1800),
	}
	d := filter.Diff(before, after)
	if len(d.Added) != 1 || d.Added[0].LeaseID != "b" {
		t.Fatalf("expected 1 added, got %+v", d.Added)
	}
	if len(d.Removed) != 0 || len(d.Changed) != 0 {
		t.Fatal("unexpected removed or changed")
	}
}

func TestDiff_Removed(t *testing.T) {
	before := []vault.SecretLease{
		makeDiffLease("a", "sec/a", "ok", 3600),
		makeDiffLease("b", "sec/b", "warn", 1800),
	}
	after := []vault.SecretLease{makeDiffLease("a", "sec/a", "ok", 3600)}
	d := filter.Diff(before, after)
	if len(d.Removed) != 1 || d.Removed[0].LeaseID != "b" {
		t.Fatalf("expected 1 removed, got %+v", d.Removed)
	}
}

func TestDiff_Changed(t *testing.T) {
	before := []vault.SecretLease{makeDiffLease("a", "sec/a", "ok", 3600)}
	after := []vault.SecretLease{makeDiffLease("a", "sec/a", "critical", 300)}
	d := filter.Diff(before, after)
	if len(d.Changed) != 1 {
		t.Fatalf("expected 1 changed, got %+v", d.Changed)
	}
}

func TestDiff_NoChange(t *testing.T) {
	leases := []vault.SecretLease{makeDiffLease("a", "sec/a", "ok", 3600)}
	d := filter.Diff(leases, leases)
	if !d.IsEmpty() {
		t.Fatalf("expected empty diff, got %s", d.Summary())
	}
}

func TestPrintDiff_ContainsStatus(t *testing.T) {
	before := []vault.SecretLease{makeDiffLease("x", "sec/x", "ok", 3600)}
	after := []vault.SecretLease{makeDiffLease("y", "sec/y", "warn", 900)}
	d := filter.Diff(before, after)
	var buf bytes.Buffer
	filter.PrintDiff(&buf, d)
	out := buf.String()
	if !strings.Contains(out, "added") || !strings.Contains(out, "removed") {
		t.Fatalf("unexpected output: %s", out)
	}
}
