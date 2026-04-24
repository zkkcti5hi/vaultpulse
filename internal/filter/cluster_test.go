package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeClusterLease(path, severity string, tags []string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "id-" + path,
		Path:      path,
		Severity:  severity,
		Tags:      tags,
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestClusterBy_Prefix(t *testing.T) {
	leases := []vault.SecretLease{
		makeClusterLease("secret/app/db", "critical", nil),
		makeClusterLease("secret/app/api", "warn", nil),
		makeClusterLease("auth/token", "ok", nil),
	}
	clusters := ClusterBy(leases, "prefix")
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	keys := map[string]int{}
	for _, c := range clusters {
		keys[c.Key] = len(c.Leases)
	}
	if keys["secret"] != 2 {
		t.Errorf("expected 2 leases under 'secret', got %d", keys["secret"])
	}
	if keys["auth"] != 1 {
		t.Errorf("expected 1 lease under 'auth', got %d", keys["auth"])
	}
}

func TestClusterBy_Severity(t *testing.T) {
	leases := []vault.SecretLease{
		makeClusterLease("a", "critical", nil),
		makeClusterLease("b", "critical", nil),
		makeClusterLease("c", "warn", nil),
	}
	clusters := ClusterBy(leases, "severity")
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	for _, c := range clusters {
		if c.Key == "critical" && len(c.Leases) != 2 {
			t.Errorf("expected 2 critical leases, got %d", len(c.Leases))
		}
	}
}

func TestClusterBy_Tag(t *testing.T) {
	leases := []vault.SecretLease{
		makeClusterLease("a", "ok", []string{"prod"}),
		makeClusterLease("b", "ok", []string{"prod"}),
		makeClusterLease("c", "ok", nil),
	}
	clusters := ClusterBy(leases, "tag")
	keys := map[string]int{}
	for _, c := range clusters {
		keys[c.Key] = len(c.Leases)
	}
	if keys["prod"] != 2 {
		t.Errorf("expected 2 'prod' leases, got %d", keys["prod"])
	}
	if keys["untagged"] != 1 {
		t.Errorf("expected 1 untagged lease, got %d", keys["untagged"])
	}
}

func TestPrintClusters_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeClusterLease("secret/db", "critical", nil),
	}
	clusters := ClusterBy(leases, "prefix")
	var buf bytes.Buffer
	PrintClusters(clusters, &buf)
	out := buf.String()
	if !containsStr(out, "CLUSTER") {
		t.Errorf("expected CLUSTER header in output, got: %s", out)
	}
	if !containsStr(out, "COUNT") {
		t.Errorf("expected COUNT header in output, got: %s", out)
	}
}

func TestPrintClusters_Empty(t *testing.T) {
	var buf bytes.Buffer
	PrintClusters(nil, &buf)
	if !containsStr(buf.String(), "no clusters") {
		t.Errorf("expected 'no clusters' message")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstring(s, sub))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
