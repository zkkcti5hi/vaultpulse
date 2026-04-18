package filter

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeSearchLease(id, path, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		Severity: severity,
		Expiry:   time.Now().Add(time.Hour),
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	leases := []vault.SecretLease{
		makeSearchLease("lease/a", "secret/foo", "ok"),
		makeSearchLease("lease/b", "secret/bar", "ok"),
	}
	out := Search(leases, SearchOptions{})
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
}

func TestSearch_MatchesPath(t *testing.T) {
	leases := []vault.SecretLease{
		makeSearchLease("lease/1", "secret/db/prod", "ok"),
		makeSearchLease("lease/2", "secret/db/staging", "ok"),
		makeSearchLease("lease/3", "secret/api/key", "ok"),
	}
	out := Search(leases, SearchOptions{Query: "db"})
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
}

func TestSearch_MatchesLeaseID(t *testing.T) {
	leases := []vault.SecretLease{
		makeSearchLease("aws/creds/role/abc123", "aws/creds", "ok"),
		makeSearchLease("database/creds/xyz", "database/creds", "ok"),
	}
	out := Search(leases, SearchOptions{Query: "abc123"})
	if len(out) != 1 {
		t.Fatalf("expected 1, got %d", len(out))
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	leases := []vault.SecretLease{
		makeSearchLease("lease/1", "secret/MyApp", "ok"),
	}
	out := Search(leases, SearchOptions{Query: "myapp"})
	if len(out) != 1 {
		t.Fatalf("expected 1, got %d", len(out))
	}
}

func TestSearch_CaseSensitiveNoMatch(t *testing.T) {
	leases := []vault.SecretLease{
		makeSearchLease("lease/1", "secret/MyApp", "ok"),
	}
	out := Search(leases, SearchOptions{Query: "myapp", CaseSensitive: true})
	if len(out) != 0 {
		t.Fatalf("expected 0, got %d", len(out))
	}
}

func TestSearch_NoMatch(t *testing.T) {
	leases := []vault.SecretLease{
		makeSearchLease("lease/1", "secret/foo", "ok"),
	}
	out := Search(leases, SearchOptions{Query: "zzznomatch"})
	if len(out) != 0 {
		t.Fatalf("expected 0, got %d", len(out))
	}
}
