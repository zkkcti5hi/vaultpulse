package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeSimilarityLease(path string, tags ...string) vault.SecretLease {
	meta := map[string][]string{}
	if len(tags) > 0 {
		meta["tags"] = tags
	}
	return vault.SecretLease{
		LeaseID:  "lease-" + path,
		Path:     path,
		ExpireAt: time.Now().Add(2 * time.Hour),
		Metadata: meta,
	}
}

func TestFindSimilar_Empty(t *testing.T) {
	pairs := FindSimilar(nil, DefaultSimilarityOptions())
	if len(pairs) != 0 {
		t.Fatalf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestFindSimilar_IdenticalPaths(t *testing.T) {
	leases := []vault.SecretLease{
		makeSimilarityLease("secret/app/db"),
		makeSimilarityLease("secret/app/db"),
	}
	pairs := FindSimilar(leases, DefaultSimilarityOptions())
	if len(pairs) == 0 {
		t.Fatal("expected at least one similar pair for identical paths")
	}
	if pairs[0].Score != 1.0 {
		t.Errorf("expected score 1.0, got %.2f", pairs[0].Score)
	}
}

func TestFindSimilar_DissimilarPaths(t *testing.T) {
	leases := []vault.SecretLease{
		makeSimilarityLease("secret/app/db"),
		makeSimilarityLease("kv/totally/different/path"),
	}
	opts := DefaultSimilarityOptions()
	opts.MinScore = 0.8
	pairs := FindSimilar(leases, opts)
	if len(pairs) != 0 {
		t.Fatalf("expected 0 pairs for dissimilar paths, got %d", len(pairs))
	}
}

func TestFindSimilar_TagBoost(t *testing.T) {
	leases := []vault.SecretLease{
		makeSimilarityLease("secret/a", "prod", "db"),
		makeSimilarityLease("secret/b", "prod", "db"),
	}
	pairs := FindSimilar(leases, DefaultSimilarityOptions())
	if len(pairs) == 0 {
		t.Fatal("expected pairs with shared tags to be similar")
	}
}

func TestFindSimilar_MaxResults(t *testing.T) {
	leases := []vault.SecretLease{
		makeSimilarityLease("secret/app/a"),
		makeSimilarityLease("secret/app/b"),
		makeSimilarityLease("secret/app/c"),
		makeSimilarityLease("secret/app/d"),
	}
	opts := DefaultSimilarityOptions()
	opts.MinScore = 0.0
	opts.MaxResults = 2
	pairs := FindSimilar(leases, opts)
	if len(pairs) > 2 {
		t.Errorf("expected at most 2 pairs, got %d", len(pairs))
	}
}

func TestPrintSimilarity_Empty(t *testing.T) {
	var buf bytes.Buffer
	PrintSimilarity(nil, &buf)
	if !bytes.Contains(buf.Bytes(), []byte("no similar")) {
		t.Error("expected 'no similar' message for empty pairs")
	}
}

func TestPrintSimilarity_ContainsHeaders(t *testing.T) {
	leases := []vault.SecretLease{
		makeSimilarityLease("secret/app/db"),
		makeSimilarityLease("secret/app/db"),
	}
	pairs := FindSimilar(leases, DefaultSimilarityOptions())
	var buf bytes.Buffer
	PrintSimilarity(pairs, &buf)
	out := buf.String()
	for _, h := range []string{"LEASE A", "LEASE B", "SCORE"} {
		if !bytes.Contains([]byte(out), []byte(h)) {
			t.Errorf("expected header %q in output", h)
		}
	}
}
