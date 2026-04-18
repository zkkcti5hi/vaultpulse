package filter_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeTagLease(path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   path + "/lease",
		Path:      path,
		ExpiresAt: time.Now().Add(time.Hour),
		Severity:  "ok",
	}
}

func TestFilterByTags_EmptyTags(t *testing.T) {
	leases := []vault.SecretLease{makeTagLease("secret/db/prod"), makeTagLease("secret/api/dev")}
	got := filter.FilterByTags(leases, nil)
	if len(got) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(got))
	}
}

func TestFilterByTags_MatchesSingle(t *testing.T) {
	leases := []vault.SecretLease{
		makeTagLease("secret/db/prod"),
		makeTagLease("secret/api/dev"),
		makeTagLease("secret/db/staging"),
	}
	got := filter.FilterByTags(leases, []string{"db"})
	if len(got) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(got))
	}
}

func TestFilterByTags_MatchesAny(t *testing.T) {
	leases := []vault.SecretLease{
		makeTagLease("secret/db/prod"),
		makeTagLease("secret/api/dev"),
		makeTagLease("secret/cache/prod"),
	}
	got := filter.FilterByTags(leases, []string{"api", "cache"})
	if len(got) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(got))
	}
}

func TestFilterByTags_CaseInsensitive(t *testing.T) {
	leases := []vault.SecretLease{makeTagLease("secret/DB/prod")}
	got := filter.FilterByTags(leases, []string{"db"})
	if len(got) != 1 {
		t.Fatalf("expected 1 lease, got %d", len(got))
	}
}

func TestFilterByTags_NoMatch(t *testing.T) {
	leases := []vault.SecretLease{makeTagLease("secret/db/prod")}
	got := filter.FilterByTags(leases, []string{"redis"})
	if len(got) != 0 {
		t.Fatalf("expected 0 leases, got %d", len(got))
	}
}
