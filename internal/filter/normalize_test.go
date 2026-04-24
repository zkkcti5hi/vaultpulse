package filter

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeNormalizeLease(path, leaseID string, meta map[string]string) vault.SecretLease {
	return vault.SecretLease{
		Path:      path,
		LeaseID:   leaseID,
		ExpiresAt: time.Now().Add(time.Hour),
		Metadata:  meta,
	}
}

func TestNormalize_TrimSpace(t *testing.T) {
	leases := []vault.SecretLease{
		makeNormalizeLease("  secret/foo  ", "  lease-1  ", nil),
	}
	opts := DefaultNormalizeOptions()
	out := Normalize(leases, opts)
	if out[0].Path != "secret/foo" {
		t.Errorf("expected trimmed path, got %q", out[0].Path)
	}
	if out[0].LeaseID != "lease-1" {
		t.Errorf("expected trimmed leaseID, got %q", out[0].LeaseID)
	}
}

func TestNormalize_LowercasePath(t *testing.T) {
	leases := []vault.SecretLease{
		makeNormalizeLease("Secret/FOO/Bar", "LEASE-1", nil),
	}
	opts := DefaultNormalizeOptions()
	opts.LowercasePath = true
	out := Normalize(leases, opts)
	if out[0].Path != "secret/foo/bar" {
		t.Errorf("expected lowercase path, got %q", out[0].Path)
	}
	// LeaseID should remain unchanged when LowercaseLeaseID is false
	if out[0].LeaseID != "LEASE-1" {
		t.Errorf("expected unchanged leaseID, got %q", out[0].LeaseID)
	}
}

func TestNormalize_LowercaseLeaseID(t *testing.T) {
	leases := []vault.SecretLease{
		makeNormalizeLease("secret/foo", "LEASE-UPPER-123", nil),
	}
	opts := DefaultNormalizeOptions()
	opts.LowercaseLeaseID = true
	out := Normalize(leases, opts)
	if out[0].LeaseID != "lease-upper-123" {
		t.Errorf("expected lowercase leaseID, got %q", out[0].LeaseID)
	}
}

func TestNormalize_CollapseMetaSpaces(t *testing.T) {
	meta := map[string]string{
		"owner": "  alice   bob  ",
		"env":   "prod",
	}
	leases := []vault.SecretLease{
		makeNormalizeLease("secret/foo", "lease-1", meta),
	}
	opts := DefaultNormalizeOptions()
	out := Normalize(leases, opts)
	if out[0].Metadata["owner"] != "alice bob" {
		t.Errorf("expected collapsed spaces, got %q", out[0].Metadata["owner"])
	}
	if out[0].Metadata["env"] != "prod" {
		t.Errorf("expected unchanged value, got %q", out[0].Metadata["env"])
	}
}

func TestNormalize_DoesNotMutateInput(t *testing.T) {
	original := makeNormalizeLease("  SECRET/FOO  ", "  LEASE-1  ", map[string]string{"k": "  v  "})
	leases := []vault.SecretLease{original}
	opts := DefaultNormalizeOptions()
	opts.LowercasePath = true
	_ = Normalize(leases, opts)
	if leases[0].Path != "  SECRET/FOO  " {
		t.Error("Normalize mutated the original slice")
	}
}

func TestNormalize_Empty(t *testing.T) {
	out := Normalize([]vault.SecretLease{}, DefaultNormalizeOptions())
	if len(out) != 0 {
		t.Errorf("expected empty result, got %d leases", len(out))
	}
}
