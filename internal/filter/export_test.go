package filter_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeExportLease(id, path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		TTL:       ttl,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestExport_CSV_Headers(t *testing.T) {
	leases := []vault.SecretLease{makeExportLease("id1", "secret/foo", "critical", time.Minute)}
	var buf bytes.Buffer
	if err := filter.Export(&buf, leases, filter.FormatCSV); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, h := range []string{"lease_id", "path", "severity", "expires_at", "ttl_seconds"} {
		if !strings.Contains(out, h) {
			t.Errorf("missing header %q", h)
		}
	}
}

func TestExport_CSV_Data(t *testing.T) {
	leases := []vault.SecretLease{makeExportLease("abc", "secret/bar", "warning", 2*time.Minute)}
	var buf bytes.Buffer
	_ = filter.Export(&buf, leases, filter.FormatCSV)
	out := buf.String()
	if !strings.Contains(out, "abc") || !strings.Contains(out, "secret/bar") {
		t.Errorf("CSV missing lease data: %s", out)
	}
}

func TestExport_JSON_Valid(t *testing.T) {
	leases := []vault.SecretLease{makeExportLease("j1", "kv/test", "ok", time.Hour)}
	var buf bytes.Buffer
	if err := filter.Export(&buf, leases, filter.FormatJSON); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "kv/test") {
		t.Error("JSON missing path")
	}
}

func TestExport_Text_Format(t *testing.T) {
	leases := []vault.SecretLease{makeExportLease("t1", "pki/cert", "critical", 30*time.Second)}
	var buf bytes.Buffer
	_ = filter.Export(&buf, leases, filter.FormatText)
	out := buf.String()
	if !strings.Contains(out, "pki/cert") || !strings.Contains(out, "critical") {
		t.Errorf("text output missing data: %s", out)
	}
}

func TestExport_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := filter.Export(&buf, nil, filter.FormatCSV); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected only header line, got %d lines", len(lines))
	}
}
