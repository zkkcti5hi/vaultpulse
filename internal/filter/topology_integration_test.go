package filter_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeIntTopologyLease(path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "int-" + path,
		Path:      path,
		TTL:       ttl,
		ExpiresAt: time.Now().Add(ttl),
		Severity:  severity,
	}
}

func TestTopology_FilterThenBuild_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntTopologyLease("secret/app/db", "critical", 10*time.Minute),
		makeIntTopologyLease("secret/app/cache", "warn", 2*time.Hour),
		makeIntTopologyLease("kv/infra/tls", "ok", 24*time.Hour),
	}

	// Filter to only critical/warn, then build topology.
	filtered := filter.Apply(leases, filter.Options{MinSeverity: "warn"})
	root := filter.BuildTopology(filtered)

	// kv/infra/tls is ok, should be excluded.
	if _, ok := root.Children["kv"]; ok {
		t.Error("expected 'kv' branch to be absent after filtering")
	}
	if _, ok := root.Children["secret"]; !ok {
		t.Error("expected 'secret' branch to be present")
	}
}

func TestTopology_EmptyLeases_Integration(t *testing.T) {
	root := filter.BuildTopology(nil)
	var buf bytes.Buffer
	filter.PrintTopology(root, &buf)
	if strings.TrimSpace(buf.String()) == "" {
		// root node "/" should still be printed
		t.Error("expected at least root node in output")
	}
}

func TestTopology_DeepPath_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntTopologyLease("secret/team/app/env/db", "critical", 5*time.Minute),
	}
	root := filter.BuildTopology(leases)
	var buf bytes.Buffer
	filter.PrintTopology(root, &buf)
	out := buf.String()
	for _, seg := range []string{"secret", "team", "app", "env", "db"} {
		if !strings.Contains(out, seg) {
			t.Errorf("expected segment %q in topology output", seg)
		}
	}
}
