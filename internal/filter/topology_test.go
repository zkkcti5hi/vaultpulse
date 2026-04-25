package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeTopologyLease(path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   "id-" + path,
		Path:      path,
		TTL:       time.Hour,
		ExpiresAt: time.Now().Add(time.Hour),
		Severity:  "ok",
	}
}

func TestBuildTopology_SingleLease(t *testing.T) {
	leases := []vault.SecretLease{makeTopologyLease("secret/app/db")}
	root := BuildTopology(leases)
	if root == nil {
		t.Fatal("expected non-nil root")
	}
	app, ok := root.Children["secret"]
	if !ok {
		t.Fatal("expected 'secret' child")
	}
	db, ok := app.Children["app"]
	if !ok {
		t.Fatal("expected 'app' child")
	}
	leaf := db.Children["db"]
	if len(leaf.Leases) != 1 {
		t.Fatalf("expected 1 lease at leaf, got %d", len(leaf.Leases))
	}
}

func TestBuildTopology_MultipleLeasesSamePath(t *testing.T) {
	leases := []vault.SecretLease{
		makeTopologyLease("secret/app"),
		makeTopologyLease("secret/app"),
	}
	root := BuildTopology(leases)
	leaf := root.Children["secret"].Children["app"]
	if len(leaf.Leases) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(leaf.Leases))
	}
}

func TestBuildTopology_Empty(t *testing.T) {
	root := BuildTopology(nil)
	if len(root.Children) != 0 {
		t.Fatalf("expected no children for empty input")
	}
}

func TestPrintTopology_ContainsPath(t *testing.T) {
	leases := []vault.SecretLease{
		makeTopologyLease("secret/db/creds"),
	}
	root := BuildTopology(leases)
	var buf bytes.Buffer
	PrintTopology(root, &buf)
	out := buf.String()
	if !strings.Contains(out, "secret") {
		t.Errorf("expected 'secret' in output, got: %s", out)
	}
	if !strings.Contains(out, "creds") {
		t.Errorf("expected 'creds' in output, got: %s", out)
	}
}

func TestPrintTopology_LeaseCount(t *testing.T) {
	leases := []vault.SecretLease{
		makeTopologyLease("kv/svc"),
		makeTopologyLease("kv/svc"),
	}
	root := BuildTopology(leases)
	var buf bytes.Buffer
	PrintTopology(root, &buf)
	if !strings.Contains(buf.String(), "2 lease(s)") {
		t.Errorf("expected lease count in output: %s", buf.String())
	}
}
