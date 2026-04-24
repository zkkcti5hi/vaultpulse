package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeLifecycleLease(id string, issuedAgo, expiresIn time.Duration, renewedAt string) vault.SecretLease {
	now := time.Now()
	meta := map[string]string{}
	if renewedAt != "" {
		meta["renewed_at"] = renewedAt
	}
	return vault.SecretLease{
		LeaseID:   id,
		Path:      "secret/" + id,
		IssuedAt:  now.Add(-issuedAgo),
		ExpiresAt: now.Add(expiresIn),
		Metadata:  meta,
	}
}

func TestClassifyLifecycle_Active(t *testing.T) {
	leases := []vault.SecretLease{
		makeLifecycleLease("a", time.Hour, 48*time.Hour, ""),
	}
	entries := ClassifyLifecycle(leases, 24*time.Hour)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Stage != StageActive {
		t.Errorf("expected active, got %s", entries[0].Stage)
	}
}

func TestClassifyLifecycle_Expiring(t *testing.T) {
	leases := []vault.SecretLease{
		makeLifecycleLease("b", time.Hour, 2*time.Hour, ""),
	}
	entries := ClassifyLifecycle(leases, 24*time.Hour)
	if entries[0].Stage != StageExpiring {
		t.Errorf("expected expiring, got %s", entries[0].Stage)
	}
}

func TestClassifyLifecycle_Expired(t *testing.T) {
	leases := []vault.SecretLease{
		makeLifecycleLease("c", 2*time.Hour, -1*time.Minute, ""),
	}
	entries := ClassifyLifecycle(leases, 24*time.Hour)
	if entries[0].Stage != StageExpired {
		t.Errorf("expected expired, got %s", entries[0].Stage)
	}
}

func TestClassifyLifecycle_Renewing(t *testing.T) {
	leases := []vault.SecretLease{
		makeLifecycleLease("d", time.Hour, 48*time.Hour, "2024-01-01T00:00:00Z"),
	}
	entries := ClassifyLifecycle(leases, 24*time.Hour)
	if entries[0].Stage != StageRenewing {
		t.Errorf("expected renewing, got %s", entries[0].Stage)
	}
}

func TestFilterByStage_ReturnsMatchingOnly(t *testing.T) {
	leases := []vault.SecretLease{
		makeLifecycleLease("e", time.Hour, 48*time.Hour, ""),
		makeLifecycleLease("f", time.Hour, 2*time.Hour, ""),
	}
	entries := ClassifyLifecycle(leases, 24*time.Hour)
	expiring := FilterByStage(entries, StageExpiring)
	if len(expiring) != 1 {
		t.Fatalf("expected 1 expiring, got %d", len(expiring))
	}
	if expiring[0].Lease.LeaseID != "f" {
		t.Errorf("unexpected lease id %s", expiring[0].Lease.LeaseID)
	}
}

func TestPrintLifecycle_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeLifecycleLease("g", time.Hour, 48*time.Hour, ""),
	}
	entries := ClassifyLifecycle(leases, 24*time.Hour)
	var buf bytes.Buffer
	PrintLifecycle(entries, &buf)
	out := buf.String()
	for _, hdr := range []string{"LEASE ID", "STAGE", "AGE", "EXPIRES AT"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("missing header %q in output", hdr)
		}
	}
}

func TestClassifyLifecycle_Empty(t *testing.T) {
	entries := ClassifyLifecycle(nil, time.Hour)
	if len(entries) != 0 {
		t.Errorf("expected empty, got %d", len(entries))
	}
}
