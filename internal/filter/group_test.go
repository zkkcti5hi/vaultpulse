package filter

import (
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

func makeGroupLease(path, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   path + "-id",
		Path:      path,
		ExpiresAt: time.Now().Add(time.Hour),
		Severity:  severity,
	}
}

func TestGroupBySeverity_Keys(t *testing.T) {
	leases := []vault.SecretLease{
		makeGroupLease("secret/a", "critical"),
		makeGroupLease("secret/b", "warning"),
		makeGroupLease("secret/c", "critical"),
	}
	got := GroupBySeverity(leases)
	if len(got["critical"]) != 2 {
		t.Errorf("expected 2 critical, got %d", len(got["critical"]))
	}
	if len(got["warning"]) != 1 {
		t.Errorf("expected 1 warning, got %d", len(got["warning"]))
	}
}

func TestGroupBySeverity_Empty(t *testing.T) {
	got := GroupBySeverity(nil)
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestGroupByPath_Prefix(t *testing.T) {
	leases := []vault.SecretLease{
		makeGroupLease("secret/foo/bar", "ok"),
		makeGroupLease("secret/baz", "ok"),
		makeGroupLease("kv/mykey", "warning"),
	}
	got := GroupByPath(leases)
	if len(got["secret"]) != 2 {
		t.Errorf("expected 2 under 'secret', got %d", len(got["secret"]))
	}
	if len(got["kv"]) != 1 {
		t.Errorf("expected 1 under 'kv', got %d", len(got["kv"]))
	}
}

func TestGroupByPath_NoSlash(t *testing.T) {
	leases := []vault.SecretLease{
		makeGroupLease("rootsecret", "ok"),
	}
	got := GroupByPath(leases)
	if _, ok := got["rootsecret"]; !ok {
		t.Error("expected key 'rootsecret'")
	}
}
