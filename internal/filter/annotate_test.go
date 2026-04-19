package filter

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeAnnotateLease(id, path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		ExpiresAt: time.Now().Add(2 * time.Hour),
		Severity:  "warning",
	}
}

func TestAnnotateLeases_TagsAndLabels(t *testing.T) {
	leases := []vault.SecretLease{
		makeAnnotateLease("lease/1", "secret/app"),
		makeAnnotateLease("lease/2", "secret/db"),
	}
	opts := AnnotateOptions{
		AddTags:   []string{"prod", "critical"},
		AddLabels: []string{"team-a"},
	}
	results := AnnotateLeases(leases, opts)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if len(results[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(results[0].Tags))
	}
	if len(results[0].Labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(results[0].Labels))
	}
}

func TestAnnotateLeases_NotePrefix(t *testing.T) {
	leases := []vault.SecretLease{makeAnnotateLease("lease/abc", "secret/x")}
	opts := AnnotateOptions{NotePrefix: "review"}
	results := AnnotateLeases(leases, opts)
	expected := "review: lease/abc"
	if results[0].Note != expected {
		t.Errorf("expected note %q, got %q", expected, results[0].Note)
	}
}

func TestAnnotateLeases_DeduplicatesTags(t *testing.T) {
	leases := []vault.SecretLease{makeAnnotateLease("lease/1", "secret/a")}
	opts := AnnotateOptions{AddTags: []string{"prod", "PROD", "Prod"}}
	results := AnnotateLeases(leases, opts)
	if len(results[0].Tags) != 1 {
		t.Errorf("expected 1 deduplicated tag, got %d", len(results[0].Tags))
	}
}

func TestAnnotateLeases_Empty(t *testing.T) {
	results := AnnotateLeases([]vault.SecretLease{}, AnnotateOptions{})
	if len(results) != 0 {
		t.Errorf("expected empty results")
	}
}

func TestAnnotateLeases_NoNote(t *testing.T) {
	leases := []vault.SecretLease{makeAnnotateLease("lease/1", "secret/b")}
	results := AnnotateLeases(leases, AnnotateOptions{})
	if results[0].Note != "" {
		t.Errorf("expected empty note, got %q", results[0].Note)
	}
}
