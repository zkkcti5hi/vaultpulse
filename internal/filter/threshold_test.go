package filter_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func makeThresholdLease(ttlHours float64) vault.SecretLease {
	return vault.SecretLease{
		LeaseID: "lease/thresh",
		Path:    "secret/thresh",
		TTL:     time.Duration(ttlHours * float64(time.Hour)),
		Severity: "ok",
	}
}

func TestApplyThreshold_Critical(t *testing.T) {
	leases := []vault.SecretLease{makeThresholdLease(10)}
	out := filter.ApplyThreshold(leases, filter.DefaultThreshold)
	if out[0].Severity != "critical" {
		t.Errorf("expected critical, got %s", out[0].Severity)
	}
}

func TestApplyThreshold_Warn(t *testing.T) {
	leases := []vault.SecretLease{makeThresholdLease(48)}
	out := filter.ApplyThreshold(leases, filter.DefaultThreshold)
	if out[0].Severity != "warn" {
		t.Errorf("expected warn, got %s", out[0].Severity)
	}
}

func TestApplyThreshold_OK(t *testing.T) {
	leases := []vault.SecretLease{makeThresholdLease(100)}
	out := filter.ApplyThreshold(leases, filter.DefaultThreshold)
	if out[0].Severity != "ok" {
		t.Errorf("expected ok, got %s", out[0].Severity)
	}
}

func TestApplyThreshold_DoesNotMutateInput(t *testing.T) {
	original := makeThresholdLease(10)
	original.Severity = "ok"
	leases := []vault.SecretLease{original}
	filter.ApplyThreshold(leases, filter.DefaultThreshold)
	if leases[0].Severity != "ok" {
		t.Error("input slice was mutated")
	}
}

func TestParseThresholdFlag_Valid(t *testing.T) {
	cfg, err := filter.ParseThresholdFlag("warn=48,critical=12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WarnHours != 48 || cfg.CriticalHours != 12 {
		t.Errorf("unexpected values: %+v", cfg)
	}
}

func TestParseThresholdFlag_Invalid(t *testing.T) {
	_, err := filter.ParseThresholdFlag("bad")
	if err == nil {
		t.Error("expected error for bad format")
	}
}

func TestParseThresholdFlag_CriticalGEWarn(t *testing.T) {
	_, err := filter.ParseThresholdFlag("warn=10,critical=20")
	if err == nil {
		t.Error("expected error when critical >= warn")
	}
}

func TestParseThresholdFlag_Empty(t *testing.T) {
	cfg, err := filter.ParseThresholdFlag("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WarnHours != filter.DefaultThreshold.WarnHours {
		t.Error("expected defaults for empty string")
	}
}
