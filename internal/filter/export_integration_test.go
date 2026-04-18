package filter_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeIntExportLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "a", Path: "secret/alpha", Severity: "critical", TTL: 10 * time.Second, ExpiresAt: now.Add(10 * time.Second)},
		{LeaseID: "b", Path: "secret/beta", Severity: "warning", TTL: time.Minute, ExpiresAt: now.Add(time.Minute)},
		{LeaseID: "c", Path: "kv/gamma", Severity: "ok", TTL: time.Hour, ExpiresAt: now.Add(time.Hour)},
	}
}

func TestExport_CSV_ParseableRoundtrip(t *testing.T) {
	leases := makeIntExportLeases()
	var buf bytes.Buffer
	if err := filter.Export(&buf, leases, filter.FormatCSV); err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	// header + 3 data rows
	if len(records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(records))
	}
}

func TestExport_JSON_ParseableRoundtrip(t *testing.T) {
	leases := makeIntExportLeases()
	var buf bytes.Buffer
	if err := filter.Export(&buf, leases, filter.FormatJSON); err != nil {
		t.Fatal(err)
	}
	var out []vault.SecretLease
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("JSON not parseable: %v", err)
	}
	if len(out) != len(leases) {
		t.Errorf("expected %d leases, got %d", len(leases), len(out))
	}
}

func TestExport_FilterThenExport_Integration(t *testing.T) {
	leases := makeIntExportLeases()
	filtered := filter.Apply(leases, filter.Options{MinSeverity: "warning"})
	var buf bytes.Buffer
	_ = filter.Export(&buf, filtered, filter.FormatText)
	out := buf.String()
	if strings.Contains(out, "kv/gamma") {
		t.Error("ok-severity lease should have been filtered out")
	}
	if !strings.Contains(out, "secret/alpha") {
		t.Error("critical lease should be present")
	}
}
