package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/vaultpulse/internal/report"
	"github.com/user/vaultpulse/internal/vault"
)

func makeLeases() []vault.Lease {
	return []vault.Lease{
		{
			LeaseID:   "lease/db/1",
			Path:      "database/creds/readonly",
			ExpiresAt: time.Now().Add(30 * time.Minute),
			Severity:  "warning",
		},
		{
			LeaseID:   "lease/aws/2",
			Path:      "aws/creds/deploy",
			ExpiresAt: time.Now().Add(5 * time.Minute),
			Severity:  "critical",
		},
	}
}

func TestReporter_Table_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	r := report.NewReporter(&buf, report.FormatTable)
	if err := r.Write(makeLeases()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, header := range []string{"PATH", "SEVERITY", "EXPIRES IN", "LEASE ID"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected header %q in output", header)
		}
	}
}

func TestReporter_Table_ContainsLeaseData(t *testing.T) {
	var buf bytes.Buffer
	r := report.NewReporter(&buf, report.FormatTable)
	_ = r.Write(makeLeases())
	out := buf.String()
	if !strings.Contains(out, "database/creds/readonly") {
		t.Error("expected lease path in table output")
	}
	if !strings.Contains(out, "critical") {
		t.Error("expected severity in table output")
	}
}

func TestReporter_JSON_ContainsLeaseData(t *testing.T) {
	var buf bytes.Buffer
	r := report.NewReporter(&buf, report.FormatJSON)
	_ = r.Write(makeLeases())
	out := buf.String()
	if !strings.Contains(out, "aws/creds/deploy") {
		t.Error("expected lease path in JSON output")
	}
	if !strings.Contains(out, "warning") {
		t.Error("expected severity in JSON output")
	}
}

func TestReporter_Defaults(t *testing.T) {
	var buf bytes.Buffer
	r := report.NewReporter(&buf, "")
	if err := r.Write([]vault.Lease{}); err != nil {
		t.Fatalf("unexpected error on empty leases: %v", err)
	}
}
