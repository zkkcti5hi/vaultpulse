package filter

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

func makePriorityLease(id, path, severity string, tags []string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Tags:      tags,
	}
}

func TestApplyPriority_NoRules(t *testing.T) {
	leases := []vault.SecretLease{
		makePriorityLease("a", "secret/foo", "ok", nil),
		makePriorityLease("b", "secret/bar", "critical", nil),
	}
	out := ApplyPriority(leases, nil)
	if out[0].LeaseID != "b" {
		t.Errorf("expected critical first, got %s", out[0].LeaseID)
	}
}

func TestApplyPriority_BoostByPath(t *testing.T) {
	leases := []vault.SecretLease{
		makePriorityLease("a", "secret/low", "warn", nil),
		makePriorityLease("b", "prod/db", "ok", nil),
	}
	rules := []PriorityRule{
		{PathPrefix: "prod/", Boost: 200},
	}
	out := ApplyPriority(leases, rules)
	if out[0].LeaseID != "b" {
		t.Errorf("expected prod/db first due to boost, got %s", out[0].LeaseID)
	}
}

func TestApplyPriority_BoostByTag(t *testing.T) {
	leases := []vault.SecretLease{
		makePriorityLease("a", "secret/x", "warn", nil),
		makePriorityLease("b", "secret/y", "ok", []string{"vip"}),
	}
	rules := []PriorityRule{
		{TagMatch: "vip", Boost: 150},
	}
	out := ApplyPriority(leases, rules)
	if out[0].LeaseID != "b" {
		t.Errorf("expected vip-tagged lease first, got %s", out[0].LeaseID)
	}
}

func TestApplyPriority_Empty(t *testing.T) {
	out := ApplyPriority(nil, nil)
	if out != nil {
		t.Error("expected nil for empty input")
	}
}

func TestApplyPriority_StableOrder(t *testing.T) {
	leases := []vault.SecretLease{
		makePriorityLease("a", "secret/a", "critical", nil),
		makePriorityLease("b", "secret/b", "critical", nil),
	}
	out := ApplyPriority(leases, nil)
	if out[0].LeaseID != "a" {
		t.Error("expected stable order preserved for equal scores")
	}
}
